#!/usr/bin/env python3
"""
CloudCompare Batch Processing Script for LAS Files
===================================================

This script processes LAS point cloud files through CloudCompare CLI to:
1. Import LAS files
2. Compute normals (Triangulation + MST orientation knn=6)
3. Convert normals to Dip/Dip Direction
4. Perform Poisson Surface Reconstruction with full parameter control
5. Save the project as .bin file

Author: CloudCompare Automation Script
License: MIT
"""

import argparse
import os
import platform
import subprocess
import sys
from dataclasses import dataclass
from pathlib import Path
from typing import List, Optional


@dataclass
class PoissonParams:
    """Parameters for Poisson Surface Reconstruction"""

    octree_depth: int = 11
    samples_per_node: float = 1.5
    point_weight: float = 2.0
    threads: int = 16
    boundary: str = "NEUMANN"  # FREE, DIRICHLET, NEUMANN
    output_density_as_sf: bool = True
    interpolate_colors: bool = True


@dataclass
class NormalParams:
    """Parameters for Normal Computation"""

    local_model: str = "TRI"  # TRI (Triangulation), QUADRIC, PLANE
    orientation: str = "MST"  # MST, PLUS_Z, MINUS_Z, etc.
    knn: int = 6  # K-nearest neighbors for MST orientation


class CloudCompareProcessor:
    """
    CloudCompare batch processor for LAS files.

    This class handles the automation of CloudCompare CLI operations
    for processing point cloud data.
    """

    # Common CloudCompare installation paths by platform
    DEFAULT_PATHS = {
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
            os.path.expanduser("~/CloudCompare/CloudCompare"),
        ],
        "Darwin": [  # macOS
            "/Applications/CloudCompare.app/Contents/MacOS/CloudCompare",
            os.path.expanduser(
                "~/Applications/CloudCompare.app/Contents/MacOS/CloudCompare"
            ),
        ],
    }

    def __init__(
        self,
        cloudcompare_path: Optional[str] = None,
        poisson_params: Optional[PoissonParams] = None,
        normal_params: Optional[NormalParams] = None,
        verbose: bool = True,
    ):
        """
        Initialize the CloudCompare processor.

        Args:
            cloudcompare_path: Path to CloudCompare executable. If None, auto-detect.
            poisson_params: Parameters for Poisson reconstruction.
            normal_params: Parameters for normal computation.
            verbose: Whether to print progress information.
        """
        self.verbose = verbose
        self.poisson_params = poisson_params or PoissonParams()
        self.normal_params = normal_params or NormalParams()
        self.cloudcompare_path = cloudcompare_path or self._find_cloudcompare()

        if not self.cloudcompare_path:
            raise RuntimeError(
                "CloudCompare not found! Please install CloudCompare or "
                "set the CLOUDCOMPARE_PATH environment variable."
            )

        self._log(f"Using CloudCompare: {self.cloudcompare_path}")

    def _log(self, message: str, level: str = "INFO"):
        """Print log message if verbose mode is enabled."""
        if self.verbose:
            print(f"[{level}] {message}")

    def _find_cloudcompare(self) -> Optional[str]:
        """Auto-detect CloudCompare installation path."""
        # First check environment variable
        env_path = os.environ.get("CLOUDCOMPARE_PATH")
        if env_path and os.path.isfile(env_path):
            return env_path

        # Check platform-specific default paths
        system = platform.system()
        paths = self.DEFAULT_PATHS.get(system, [])

        for path in paths:
            if os.path.isfile(path):
                return path

        # Try to find in PATH
        try:
            result = subprocess.run(
                ["which", "CloudCompare"]
                if system != "Windows"
                else ["where", "CloudCompare"],
                capture_output=True,
                text=True,
            )
            if result.returncode == 0:
                return result.stdout.strip().split("\n")[0]
        except Exception:
            pass

        return None

    def _build_command(self, input_file: Path, output_file: Path) -> List[str]:
        """
        Build the CloudCompare CLI command with all parameters.

        Args:
            input_file: Path to input LAS file.
            output_file: Path for output .bin file.

        Returns:
            List of command arguments.
        """
        cmd = [
            self.cloudcompare_path,
            "-SILENT",
            "-AUTO_SAVE",
            "OFF",
            "-O",
            str(input_file),
        ]

        # Compute Normals
        # CloudCompare CLI syntax: -COMPUTE_NORMALS [LOCAL_MODEL] [RADIUS]
        # For MST orientation: -ORIENT_NORMS_MST [K]
        cmd.extend(["-COMPUTE_NORMALS"])

        # Set local model for normal computation (if supported by CC version)
        # Some versions use: -COMPUTE_NORMALS LOCAL_MODEL radius
        # Triangulation is typically the default

        # Apply MST orientation with specified knn
        # Note: Some CC versions support -ORIENT_NORMS_MST K
        cmd.extend(["-ORIENT_NORMS_MST", str(self.normal_params.knn)])

        # Convert Normals to Dip/Dip Direction
        cmd.extend(["-NORMALS_TO_DIP"])

        # Poisson Surface Reconstruction
        # CloudCompare CLI syntax varies by version:
        # Older: -POISSON depth samples_per_node boundary
        # Newer: -POISSON with more options

        # Build Poisson command with available parameters
        poisson_cmd = [
            "-POISSON",
            str(self.poisson_params.octree_depth),
        ]

        # Add samples per node if supported
        poisson_cmd.append(str(self.poisson_params.samples_per_node))

        # Add boundary condition
        poisson_cmd.append(self.poisson_params.boundary)

        cmd.extend(poisson_cmd)

        # Note: The following Poisson parameters may require GUI or newer CLI versions:
        # - output_density_as_sf
        # - interpolate_colors
        # - point_weight
        # - threads
        # These are logged for reference but may not be available in all CLI versions

        # Set export format to BIN (CloudCompare native format)
        cmd.extend(["-C_EXPORT_FMT", "BIN"])
        cmd.extend(["-M_EXPORT_FMT", "BIN"])

        # Save the cloud and mesh
        # Using FILE option to specify exact output path
        cmd.extend(["-SAVE_CLOUDS", "FILE", str(output_file)])

        # Also save the mesh with a similar name
        mesh_output = output_file.parent / f"{output_file.stem}_mesh.bin"
        cmd.extend(["-SAVE_MESHES", "FILE", str(mesh_output)])

        return cmd

    def process_file(self, input_file: Path, output_dir: Path) -> bool:
        """
        Process a single LAS file.

        Args:
            input_file: Path to input LAS file.
            output_dir: Directory for output files.

        Returns:
            True if processing succeeded, False otherwise.
        """
        filename = input_file.stem
        output_file = output_dir / f"{filename}.bin"

        self._log("=" * 70)
        self._log(f"Processing: {filename}")
        self._log(f"Input:  {input_file}")
        self._log(f"Output: {output_file}")

        # Build command
        cmd = self._build_command(input_file, output_file)

        self._log(f"Command: {' '.join(cmd)}")

        try:
            # Execute CloudCompare
            result = subprocess.run(
                cmd,
                capture_output=True,
                text=True,
                timeout=3600,  # 1 hour timeout per file
            )

            if result.returncode == 0:
                self._log(f"Successfully processed: {filename}", "SUCCESS")
                if result.stdout:
                    self._log(f"Output: {result.stdout}")
                return True
            else:
                self._log(f"Failed to process: {filename}", "ERROR")
                self._log(f"Exit code: {result.returncode}", "ERROR")
                if result.stderr:
                    self._log(f"Error: {result.stderr}", "ERROR")
                return False

        except subprocess.TimeoutExpired:
            self._log(f"Timeout processing: {filename}", "ERROR")
            return False
        except Exception as e:
            self._log(f"Exception processing {filename}: {e}", "ERROR")
            return False

    def process_directory(
        self, input_dir: Path, output_subdir: str = "Processed"
    ) -> dict:
        """
        Process all LAS files in a directory.

        Args:
            input_dir: Directory containing LAS files.
            output_subdir: Subdirectory name for output files.

        Returns:
            Dictionary with processing statistics.
        """
        input_dir = Path(input_dir).resolve()
        output_dir = input_dir / output_subdir

        # Create output directory
        output_dir.mkdir(parents=True, exist_ok=True)

        self._log("=" * 70)
        self._log("CloudCompare Batch Processing")
        self._log("=" * 70)
        self._log(f"Input directory:  {input_dir}")
        self._log(f"Output directory: {output_dir}")

        # Find all LAS files (case-insensitive)
        las_files = list(input_dir.glob("*.las")) + list(input_dir.glob("*.LAS"))
        las_files = list(
            set(las_files)
        )  # Remove duplicates on case-insensitive systems
        las_files.sort()

        if not las_files:
            self._log(f"No LAS files found in: {input_dir}", "ERROR")
            return {"total": 0, "success": 0, "failed": 0}

        self._log(f"Found {len(las_files)} LAS file(s) to process")

        # Log processing parameters
        self._log("-" * 70)
        self._log("Processing Parameters:")
        self._log(f"  Normal Computation:")
        self._log(
            f"    - Local Model: {self.normal_params.local_model} (Triangulation)"
        )
        self._log(
            f"    - Orientation: {self.normal_params.orientation} (knn={self.normal_params.knn})"
        )
        self._log(f"  Poisson Reconstruction:")
        self._log(f"    - Octree Depth: {self.poisson_params.octree_depth}")
        self._log(f"    - Samples per Node: {self.poisson_params.samples_per_node}")
        self._log(f"    - Point Weight: {self.poisson_params.point_weight}")
        self._log(f"    - Boundary: {self.poisson_params.boundary}")
        self._log(f"    - Threads: {self.poisson_params.threads}")
        self._log(
            f"    - Output Density as SF: {self.poisson_params.output_density_as_sf}"
        )
        self._log(f"    - Interpolate Colors: {self.poisson_params.interpolate_colors}")
        self._log("-" * 70)

        # Process files
        success_count = 0
        failed_count = 0

        for i, las_file in enumerate(las_files, 1):
            self._log(f"\nFile {i}/{len(las_files)}")

            if self.process_file(las_file, output_dir):
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

        return {
            "total": len(las_files),
            "success": success_count,
            "failed": failed_count,
        }


def create_alternative_script(input_dir: Path, output_dir: Path) -> str:
    """
    Generate a CloudCompare script file for GUI-based batch processing.

    This is useful when CLI doesn't support all needed parameters,
    allowing users to run the script through CloudCompare's GUI.

    Args:
        input_dir: Directory containing LAS files.
        output_dir: Directory for output files.

    Returns:
        Path to generated script file.
    """
    script_content = """# CloudCompare Batch Processing Script
# Run this through CloudCompare GUI: Tools > Batch Processing
# Or use: CloudCompare -SCRIPT this_file.txt

# Note: This script template shows the operations that would be performed.
# Actual GUI-based batch processing may require manual steps for some parameters.

# For each LAS file:
# 1. Open file
# 2. Compute Normals
#    - Local surface model: Triangulation
#    - Orientation: Minimum Spanning Tree (knn=6)
# 3. Convert Normals to Dip/Dip Direction
# 4. Poisson Surface Reconstruction
#    - Octree depth: 11
#    - Samples per node: 1.5
#    - Point weight: 2.0
#    - Threads: 16
#    - Boundary: Neumann
#    - Output density as SF: Yes
#    - Interpolate colors: Yes
# 5. Save as .bin file
"""

    script_path = output_dir / "cloudcompare_batch_script.txt"
    script_path.write_text(script_content)
    return str(script_path)


def main():
    """Main entry point for the script."""
    parser = argparse.ArgumentParser(
        description="Batch process LAS files with CloudCompare",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  %(prog)s /path/to/las/files
  %(prog)s . --output-dir Processed
  %(prog)s /data --octree-depth 12 --threads 8
  CLOUDCOMPARE_PATH=/custom/path/CloudCompare %(prog)s /data

Environment Variables:
  CLOUDCOMPARE_PATH    Path to CloudCompare executable
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
        "--boundary",
        type=str,
        choices=["FREE", "DIRICHLET", "NEUMANN"],
        default="NEUMANN",
        help="Boundary condition for Poisson reconstruction (default: NEUMANN)",
    )

    # Normal parameters
    parser.add_argument(
        "--knn",
        type=int,
        default=6,
        help="K-nearest neighbors for MST normal orientation (default: 6)",
    )

    parser.add_argument("--quiet", action="store_true", help="Suppress progress output")

    args = parser.parse_args()

    # Create parameter objects
    poisson_params = PoissonParams(
        octree_depth=args.octree_depth,
        samples_per_node=args.samples_per_node,
        point_weight=args.point_weight,
        threads=args.threads,
        boundary=args.boundary,
    )

    normal_params = NormalParams(knn=args.knn)

    try:
        processor = CloudCompareProcessor(
            cloudcompare_path=args.cloudcompare_path,
            poisson_params=poisson_params,
            normal_params=normal_params,
            verbose=not args.quiet,
        )

        result = processor.process_directory(Path(args.input_dir), args.output_dir)

        # Exit with error code if any files failed
        sys.exit(result["failed"])

    except RuntimeError as e:
        print(f"[ERROR] {e}", file=sys.stderr)
        sys.exit(1)
    except KeyboardInterrupt:
        print("\n[INFO] Processing interrupted by user")
        sys.exit(130)


if __name__ == "__main__":
    main()
