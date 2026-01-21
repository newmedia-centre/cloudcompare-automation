#!/usr/bin/env python3
"""
CloudCompare + PoissonRecon Batch Processing Script for LAS Files
=================================================================

This script processes LAS point cloud files through:
1. CloudCompare CLI - Import LAS, compute normals, convert to DIP, export PLY
2. PoissonRecon - Poisson Surface Reconstruction with density scalar field
3. CloudCompare CLI - Import point cloud and mesh into project for manual
                      filtering by density SF value

The density scalar field from PoissonRecon allows you to filter out
low-confidence areas of the reconstruction in CloudCompare by adjusting
the SF display min value and exporting the filtered mesh.

Prerequisites:
- CloudCompare (with CLI support)
- PoissonRecon from https://github.com/mkazhdan/PoissonRecon

Author: CloudCompare Automation Script
License: MIT
"""

import argparse
import os
import platform
import shutil
import subprocess
import sys
import tempfile
from dataclasses import dataclass
from pathlib import Path
from typing import List, Optional, Tuple


@dataclass
class NormalParams:
    """Parameters for Normal Computation using Octree"""

    radius: str = "AUTO"
    knn: int = 6


@dataclass
class PoissonParams:
    """Parameters for Poisson Surface Reconstruction"""

    octree_depth: int = 11
    samples_per_node: float = 1.5
    point_weight: float = 2.0
    threads: int = 16
    boundary_type: int = 3  # 1=Free, 2=Dirichlet, 3=Neumann
    output_density: bool = True  # Required for filtering in CloudCompare
    interpolate_colors: bool = True


class CloudComparePoissonProcessor:
    """
    CloudCompare + PoissonRecon batch processor for LAS files.

    Processes LAS files and outputs CloudCompare projects (.bin) containing
    both the point cloud and reconstructed mesh with density scalar field
    for manual filtering.
    """

    CLOUDCOMPARE_PATHS = {
        "Windows": [
            r"C:\Program Files\CloudCompare\CloudCompare.exe",
            r"C:\Program Files (x86)\CloudCompare\CloudCompare.exe",
            r"C:\CloudCompare\CloudCompare.exe",
        ],
        "Linux": [
            "/usr/bin/CloudCompare",
            "/usr/bin/cloudcompare",
            "/usr/local/bin/CloudCompare",
            "/usr/local/bin/cloudcompare",
            "/opt/CloudCompare/CloudCompare",
            "/snap/bin/cloudcompare",
        ],
        "Darwin": [
            "/Applications/CloudCompare.app/Contents/MacOS/CloudCompare",
        ],
    }

    POISSONRECON_PATHS = {
        "Windows": [
            r"C:\PoissonRecon\PoissonRecon.exe",
            r"C:\Program Files\PoissonRecon\PoissonRecon.exe",
        ],
        "Linux": [
            "/usr/local/bin/PoissonRecon",
            "/usr/bin/PoissonRecon",
            "/opt/PoissonRecon/PoissonRecon",
        ],
        "Darwin": [
            "/usr/local/bin/PoissonRecon",
        ],
    }

    def __init__(
        self,
        cloudcompare_path: Optional[str] = None,
        poissonrecon_path: Optional[str] = None,
        normal_params: Optional[NormalParams] = None,
        poisson_params: Optional[PoissonParams] = None,
        verbose: bool = True,
    ):
        self.verbose = verbose
        self.normal_params = normal_params or NormalParams()
        self.poisson_params = poisson_params or PoissonParams()

        self.cloudcompare_path = cloudcompare_path or self._find_executable(
            "CLOUDCOMPARE_PATH", self.CLOUDCOMPARE_PATHS, "CloudCompare"
        )
        self.poissonrecon_path = poissonrecon_path or self._find_executable(
            "POISSONRECON_PATH", self.POISSONRECON_PATHS, "PoissonRecon"
        )

        if not self.cloudcompare_path:
            raise RuntimeError(
                "CloudCompare not found! Please install CloudCompare or "
                "set the CLOUDCOMPARE_PATH environment variable."
            )
        if not self.poissonrecon_path:
            raise RuntimeError(
                "PoissonRecon not found! Please install PoissonRecon from "
                "https://github.com/mkazhdan/PoissonRecon or "
                "set the POISSONRECON_PATH environment variable."
            )

        self._log(f"Using CloudCompare: {self.cloudcompare_path}")
        self._log(f"Using PoissonRecon: {self.poissonrecon_path}")

    def _log(self, message: str, level: str = "INFO"):
        if self.verbose:
            print(f"[{level}] {message}")

    def _find_executable(
        self, env_var: str, paths_dict: dict, name: str
    ) -> Optional[str]:
        """Find executable from environment variable or common paths."""
        env_path = os.environ.get(env_var)
        if env_path and os.path.isfile(env_path):
            return env_path

        system = platform.system()
        paths = paths_dict.get(system, [])

        for path in paths:
            expanded = os.path.expanduser(path)
            if os.path.isfile(expanded):
                return expanded

        # Try to find in PATH
        result = shutil.which(name)
        if result:
            return result

        # Try lowercase on Linux
        if system == "Linux":
            result = shutil.which(name.lower())
            if result:
                return result

        return None

    def _run_command(
        self, cmd: List[str], description: str, timeout: int = 3600
    ) -> Tuple[bool, str]:
        """Run a command and return success status and output."""
        self._log(f"Running: {' '.join(cmd)}")

        try:
            result = subprocess.run(
                cmd,
                capture_output=True,
                text=True,
                timeout=timeout,
            )

            output = result.stdout + result.stderr

            if result.returncode == 0:
                return True, output
            else:
                self._log(
                    f"{description} failed with exit code {result.returncode}", "ERROR"
                )
                if output:
                    self._log(f"Output: {output}", "ERROR")
                return False, output

        except subprocess.TimeoutExpired:
            self._log(f"{description} timed out after {timeout}s", "ERROR")
            return False, "Timeout"
        except Exception as e:
            self._log(f"{description} failed with exception: {e}", "ERROR")
            return False, str(e)

    def _step1_compute_normals(self, input_las: Path, output_ply: Path) -> bool:
        """Step 1: CloudCompare - Load LAS, compute normals, convert to DIP, export PLY."""
        self._log("")
        self._log(
            "Step 1: Computing normals and converting to DIP with CloudCompare..."
        )
        self._log(f"  - Octree Radius: {self.normal_params.radius}")
        self._log(f"  - MST Orientation KNN: {self.normal_params.knn}")

        cmd = [
            self.cloudcompare_path,
            "-SILENT",
            "-AUTO_SAVE",
            "OFF",
            "-O",
            str(input_las),
            "-OCTREE_NORMALS",
            str(self.normal_params.radius),
            "-ORIENT_NORMS_MST",
            str(self.normal_params.knn),
            "-NORMALS_TO_DIP",
            "-C_EXPORT_FMT",
            "PLY",
            "-PLY_EXPORT_FMT",
            "ASCII",
            "-SAVE_CLOUDS",
            "FILE",
            str(output_ply),
        ]

        success, _ = self._run_command(cmd, "Normal computation")

        if success and output_ply.exists():
            self._log("Generated PLY with normals", "SUCCESS")
            return True
        else:
            self._log(f"Failed to generate PLY with normals: {output_ply}", "ERROR")
            return False

    def _step2_poisson_reconstruction(self, input_ply: Path, output_mesh: Path) -> bool:
        """Step 2: PoissonRecon - Poisson Surface Reconstruction with density SF."""
        self._log("")
        self._log("Step 2: Running Poisson Surface Reconstruction...")
        self._log(f"  - Octree depth: {self.poisson_params.octree_depth}")
        self._log(f"  - Samples per node: {self.poisson_params.samples_per_node}")
        self._log(f"  - Point weight: {self.poisson_params.point_weight}")
        self._log(f"  - Boundary type: {self.poisson_params.boundary_type} (Neumann)")
        self._log(f"  - Threads: {self.poisson_params.threads}")
        self._log(f"  - Output density as SF: YES (for filtering in CloudCompare)")

        cmd = [
            self.poissonrecon_path,
            "--in",
            str(input_ply),
            "--out",
            str(output_mesh),
            "--depth",
            str(self.poisson_params.octree_depth),
            "--samplesPerNode",
            str(self.poisson_params.samples_per_node),
            "--pointWeight",
            str(self.poisson_params.point_weight),
            "--bType",
            str(self.poisson_params.boundary_type),
            "--threads",
            str(self.poisson_params.threads),
        ]

        # Always output density for filtering in CloudCompare
        if self.poisson_params.output_density:
            cmd.append("--density")

        if self.poisson_params.interpolate_colors:
            cmd.append("--colors")

        success, _ = self._run_command(cmd, "Poisson reconstruction")

        if success and output_mesh.exists():
            self._log("Generated mesh with density scalar field", "SUCCESS")
            return True
        else:
            self._log(f"Failed to generate mesh: {output_mesh}", "ERROR")
            return False

    def _step3_create_project(
        self, input_ply: Path, input_mesh: Path, output_bin: Path
    ) -> bool:
        """Step 3: CloudCompare - Import point cloud and mesh, save as .bin project."""
        self._log("")
        self._log(
            "Step 3: Creating CloudCompare project (.bin) for manual filtering..."
        )
        self._log("  The mesh contains a 'Density' scalar field from PoissonRecon.")
        self._log("  In CloudCompare, you can:")
        self._log("    1. Select the mesh")
        self._log(
            "    2. Display the Density SF (Edit > Scalar Fields > Set Active SF)"
        )
        self._log("    3. Adjust SF display range to visualize low-density areas")
        self._log("    4. Filter by SF value (Edit > Scalar Fields > Filter by Value)")
        self._log("    5. Export the filtered mesh")

        cmd = [
            self.cloudcompare_path,
            "-SILENT",
            "-AUTO_SAVE",
            "OFF",
            "-O",
            str(input_ply),
            "-O",
            str(input_mesh),
            "-SAVE_CLOUDS",
            "ALL",
            "FILE",
            str(output_bin),
        ]

        success, _ = self._run_command(cmd, "Project creation")

        if success and output_bin.exists():
            self._log(f"Created CloudCompare project: {output_bin}", "SUCCESS")
            return True
        else:
            self._log(f"Failed to create project: {output_bin}", "ERROR")
            return False

    def process_file(self, input_file: Path, output_dir: Path, temp_dir: Path) -> bool:
        """Process a single LAS file through the complete pipeline."""
        filename = input_file.stem

        # Define paths
        ply_with_normals = temp_dir / f"{filename}_normals.ply"
        mesh_ply = temp_dir / f"{filename}_mesh.ply"
        output_bin = output_dir / f"{filename}.bin"

        self._log("=" * 70)
        self._log(f"Processing: {filename}")
        self._log(f"Input:  {input_file}")
        self._log(f"Output: {output_bin}")

        # Step 1: Compute normals and export PLY
        if not self._step1_compute_normals(input_file, ply_with_normals):
            return False

        # Step 2: Poisson Surface Reconstruction with density SF
        if not self._step2_poisson_reconstruction(ply_with_normals, mesh_ply):
            return False

        # Step 3: Create CloudCompare project for manual filtering
        if not self._step3_create_project(ply_with_normals, mesh_ply, output_bin):
            return False

        self._log("")
        self._log(f"Successfully processed: {filename}", "SUCCESS")
        self._log("")
        self._log("Next steps in CloudCompare GUI:")
        self._log(f"  1. Open {output_bin}")
        self._log("  2. Select the mesh in DB Tree")
        self._log("  3. Properties panel > SF Display > adjust 'displayed' min value")
        self._log(
            "  4. Use 'Edit > Scalar Fields > Filter by Value' to remove low-density vertices"
        )
        self._log("  5. Export filtered mesh via 'File > Save'")

        return True

    def process_directory(
        self, input_dir: Path, output_subdir: str = "Processed"
    ) -> dict:
        """Process all LAS files in a directory."""
        input_dir = Path(input_dir).resolve()
        output_dir = input_dir / output_subdir

        # Create output directory
        output_dir.mkdir(parents=True, exist_ok=True)

        self._log("=" * 70)
        self._log("CloudCompare + PoissonRecon Batch Processing")
        self._log("=" * 70)
        self._log(f"Input directory:  {input_dir}")
        self._log(f"Output directory: {output_dir}")

        # Find all LAS files
        las_files = list(input_dir.glob("*.las")) + list(input_dir.glob("*.LAS"))
        las_files = list(set(las_files))
        las_files.sort()

        if not las_files:
            self._log(f"No LAS files found in: {input_dir}", "ERROR")
            return {"total": 0, "success": 0, "failed": 0}

        self._log(f"Found {len(las_files)} LAS file(s) to process")

        # Log parameters
        self._log("-" * 70)
        self._log("Processing Parameters:")
        self._log("  Normal Computation:")
        self._log(f"    - Octree Radius: {self.normal_params.radius}")
        self._log(f"    - MST Orientation KNN: {self.normal_params.knn}")
        self._log("  Poisson Reconstruction:")
        self._log(f"    - Octree Depth: {self.poisson_params.octree_depth}")
        self._log(f"    - Samples per Node: {self.poisson_params.samples_per_node}")
        self._log(f"    - Point Weight: {self.poisson_params.point_weight}")
        self._log(f"    - Boundary: Neumann ({self.poisson_params.boundary_type})")
        self._log(f"    - Threads: {self.poisson_params.threads}")
        self._log(f"    - Output Density as SF: YES")
        self._log(f"    - Interpolate Colors: {self.poisson_params.interpolate_colors}")
        self._log("-" * 70)

        # Create temporary directory
        with tempfile.TemporaryDirectory(prefix="cc_poisson_") as temp_dir:
            temp_path = Path(temp_dir)
            self._log(f"Temporary directory: {temp_path}")

            # Process files
            success_count = 0
            failed_count = 0

            for i, las_file in enumerate(las_files, 1):
                self._log(f"\nFile {i}/{len(las_files)}")

                if self.process_file(las_file, output_dir, temp_path):
                    success_count += 1
                else:
                    failed_count += 1

        # Summary
        self._log("\n" + "=" * 70)
        self._log("Processing Complete")
        self._log("=" * 70)
        self._log(f"Total files:      {len(las_files)}")
        self._log(f"Successful:       {success_count}")
        self._log(f"Failed:           {failed_count}")
        self._log("")
        self._log(f"Output files are in: {output_dir}")
        self._log("  - [filename].bin : CloudCompare project with point cloud and mesh")
        self._log("")
        self._log("The mesh includes a 'Density' scalar field from PoissonRecon.")
        self._log("Open the .bin file in CloudCompare to filter out low-density areas:")
        self._log("  1. Select mesh > Properties > SF Display > adjust min value")
        self._log("  2. Edit > Scalar Fields > Filter by Value")
        self._log("  3. Export the filtered mesh")

        return {
            "total": len(las_files),
            "success": success_count,
            "failed": failed_count,
        }


def main():
    """Main entry point."""
    parser = argparse.ArgumentParser(
        description="Batch process LAS files with CloudCompare and PoissonRecon",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  %(prog)s /path/to/las/files
  %(prog)s . --output-dir Processed
  %(prog)s /data --octree-depth 12 --threads 8

Environment Variables:
  CLOUDCOMPARE_PATH    Path to CloudCompare executable
  POISSONRECON_PATH    Path to PoissonRecon executable

Prerequisites:
  - CloudCompare: https://www.cloudcompare.org/
  - PoissonRecon: https://github.com/mkazhdan/PoissonRecon

Output:
  The script creates CloudCompare project files (.bin) containing:
  - Point cloud with computed normals and DIP values
  - Reconstructed mesh with density scalar field

  The density SF allows manual filtering in CloudCompare:
  1. Open .bin file in CloudCompare
  2. Select mesh > Properties > SF Display > adjust min value
  3. Edit > Scalar Fields > Filter by Value
  4. Export the filtered mesh
        """,
    )

    parser.add_argument(
        "input_dir",
        type=str,
        nargs="?",
        default=".",
        help="Directory containing LAS files (default: current directory)",
    )

    parser.add_argument(
        "--output-dir",
        type=str,
        default="Processed",
        help="Subdirectory name for output files (default: Processed)",
    )

    parser.add_argument(
        "--cloudcompare-path",
        type=str,
        default=None,
        help="Path to CloudCompare executable",
    )

    parser.add_argument(
        "--poissonrecon-path",
        type=str,
        default=None,
        help="Path to PoissonRecon executable",
    )

    # Normal computation parameters
    parser.add_argument(
        "--radius",
        type=str,
        default="AUTO",
        help="Radius for octree normal computation (default: AUTO)",
    )

    parser.add_argument(
        "--knn",
        type=int,
        default=6,
        help="K-nearest neighbors for MST normal orientation (default: 6)",
    )

    # Poisson parameters
    parser.add_argument(
        "--octree-depth",
        type=int,
        default=11,
        help="Octree depth for Poisson reconstruction (default: 11)",
    )

    parser.add_argument(
        "--samples-per-node",
        type=float,
        default=1.5,
        help="Samples per node for Poisson reconstruction (default: 1.5)",
    )

    parser.add_argument(
        "--point-weight",
        type=float,
        default=2.0,
        help="Point weight for Poisson reconstruction (default: 2.0)",
    )

    parser.add_argument(
        "--threads",
        type=int,
        default=16,
        help="Number of threads for Poisson reconstruction (default: 16)",
    )

    parser.add_argument(
        "--boundary-type",
        type=int,
        choices=[1, 2, 3],
        default=3,
        help="Boundary type: 1=Free, 2=Dirichlet, 3=Neumann (default: 3)",
    )

    parser.add_argument(
        "--no-colors",
        action="store_true",
        help="Disable color interpolation in Poisson reconstruction",
    )

    parser.add_argument(
        "--quiet",
        action="store_true",
        help="Suppress progress output",
    )

    args = parser.parse_args()

    # Create parameter objects
    normal_params = NormalParams(
        radius=args.radius,
        knn=args.knn,
    )

    poisson_params = PoissonParams(
        octree_depth=args.octree_depth,
        samples_per_node=args.samples_per_node,
        point_weight=args.point_weight,
        threads=args.threads,
        boundary_type=args.boundary_type,
        output_density=True,  # Always enable for filtering workflow
        interpolate_colors=not args.no_colors,
    )

    try:
        processor = CloudComparePoissonProcessor(
            cloudcompare_path=args.cloudcompare_path,
            poissonrecon_path=args.poissonrecon_path,
            normal_params=normal_params,
            poisson_params=poisson_params,
            verbose=not args.quiet,
        )

        result = processor.process_directory(Path(args.input_dir), args.output_dir)
        sys.exit(result["failed"])

    except RuntimeError as e:
        print(f"[ERROR] {e}", file=sys.stderr)
        sys.exit(1)
    except KeyboardInterrupt:
        print("\n[INFO] Processing interrupted by user")
        sys.exit(130)


if __name__ == "__main__":
    main()
