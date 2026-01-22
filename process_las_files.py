#!/usr/bin/env python3
"""
CloudComPy + PoissonRecon Batch Processing Script for LAS Files
===============================================================

This script processes LAS point cloud files through CloudComPy:
1. Load LAS file
2. Compute normals using triangulation model with MST orientation
3. Convert normals to DIP/Dip Direction scalar fields
4. Run Poisson Surface Reconstruction with density scalar field
5. Save both cloud and mesh to a single .bin file

Prerequisites:
- CloudComPy (Python bindings for CloudCompare)
  https://github.com/CloudCompare/CloudComPy

Author: CloudCompare Automation Script
License: MIT
"""

import argparse
import sys
from dataclasses import dataclass
from pathlib import Path
from typing import Optional


@dataclass
class NormalParams:
    """Parameters for Normal Computation"""

    knn: int = 6  # K-nearest neighbors for MST orientation
    radius: float = 0.0  # Auto-compute radius if 0


@dataclass
class PoissonParams:
    """Parameters for Poisson Surface Reconstruction"""

    octree_depth: int = 11
    samples_per_node: float = 1.5
    point_weight: float = 2.0
    boundary_type: int = 2  # 0=FREE, 1=DIRICHLET, 2=NEUMANN


class CloudComPyProcessor:
    """
    CloudComPy batch processor for LAS files.

    Processes LAS files and outputs CloudCompare projects (.bin) containing
    both the point cloud (with normals and DIP SFs) and reconstructed mesh
    (with density scalar field).
    """

    def __init__(
        self,
        normal_params: Optional[NormalParams] = None,
        poisson_params: Optional[PoissonParams] = None,
        verbose: bool = True,
    ):
        self.verbose = verbose
        self.normal_params = normal_params or NormalParams()
        self.poisson_params = poisson_params or PoissonParams()

        # Initialize CloudComPy
        self._init_cloudcompy()

    def _log(self, message: str, level: str = "INFO"):
        if self.verbose:
            print(f"[{level}] {message}", flush=True)

    def _log_step(self, step: int, total: int, message: str):
        """Log a processing step with flush for real-time output."""
        if self.verbose:
            print(f"[INFO] [{step}/{total}] {message}", flush=True)

    def _init_cloudcompy(self):
        """Initialize CloudComPy and check for PoissonRecon plugin."""
        try:
            import cloudComPy as cc

            self.cc = cc
            self._log("CloudComPy initialized")
        except ImportError as e:
            raise RuntimeError(
                "CloudComPy not found! Please install CloudComPy.\n"
                "See: https://github.com/CloudCompare/CloudComPy\n"
                f"Error: {e}"
            )

        # Check for PoissonRecon plugin
        if not cc.isPluginPoissonRecon():
            raise RuntimeError(
                "PoissonRecon plugin not available in CloudComPy!\n"
                "Make sure CloudComPy was built with PoissonRecon support."
            )

        import cloudComPy.PoissonRecon

        self.PoissonRecon = cloudComPy.PoissonRecon
        self._log("PoissonRecon plugin loaded")

    def process_file(self, input_file: Path, output_file: Path) -> bool:
        """Process a single LAS file through the complete pipeline."""
        cc = self.cc

        self._log("=" * 70)
        self._log(f"Processing: {input_file.name}")
        self._log(f"Output: {output_file}")

        # Step 1: Load point cloud
        self._log_step(1, 5, "Loading point cloud...")
        cloud = cc.loadPointCloud(str(input_file))
        if cloud is None:
            self._log(f"Failed to load: {input_file}", "ERROR")
            return False
        self._log(f"Loaded {cloud.size():,} points", "SUCCESS")

        # Step 2: Compute normals
        self._log_step(2, 5, "Computing normals (this may take a few minutes)...")
        success = cc.computeNormals(
            [cloud],
            model=cc.LOCAL_MODEL_TYPES.TRI,  # Triangulation
            useScanGridsForComputation=False,
            defaultRadius=self.normal_params.radius,
            orientNormals=True,
            useScanGridsForOrientation=False,
            useSensorsForOrientation=False,
            orientNormalsMST=True,
            mstNeighbors=self.normal_params.knn,
        )
        if not success:
            self._log("Failed to compute normals", "ERROR")
            return False
        self._log("Normals computed", "SUCCESS")

        # Step 3: Convert normals to DIP/Dip Direction
        self._log_step(3, 5, "Converting normals to DIP/Dip Direction...")
        success = cloud.convertNormalToDipDirSFs()
        if not success:
            self._log("Failed to convert normals to DIP", "ERROR")
            return False
        self._log("DIP scalar fields created", "SUCCESS")

        # Step 4: Poisson Surface Reconstruction
        depth = self.poisson_params.octree_depth
        self._log_step(4, 5, f"Poisson Reconstruction (depth={depth})...")
        self._log(
            f"This step can take 5-30+ minutes depending on point count and depth"
        )

        # Map boundary type to enum
        boundary_map = {
            0: self.PoissonRecon.BoundaryType.FREE,
            1: self.PoissonRecon.BoundaryType.DIRICHLET,
            2: self.PoissonRecon.BoundaryType.NEUMANN,
        }
        boundary = boundary_map.get(
            self.poisson_params.boundary_type, self.PoissonRecon.BoundaryType.NEUMANN
        )

        mesh = self.PoissonRecon.PR.PoissonReconstruction(
            cloud,
            depth=self.poisson_params.octree_depth,
            samplesPerNode=self.poisson_params.samples_per_node,
            pointWeight=self.poisson_params.point_weight,
            density=True,  # Output density SF for filtering
            boundary=boundary,
        )
        if mesh is None:
            self._log("Failed to create mesh", "ERROR")
            return False
        self._log(f"Mesh created with {mesh.size():,} faces", "SUCCESS")

        # Step 4b: Transfer colors from source cloud to mesh vertices
        if cloud.hasColors():
            self._log("Transferring colors to mesh...")
            mesh_cloud = mesh.getAssociatedCloud()
            if mesh_cloud is not None:
                success = mesh_cloud.interpolateColorsFrom(cloud)
                if success:
                    self._log("Colors transferred to mesh", "SUCCESS")
                else:
                    self._log("Failed to interpolate colors", "WARNING")
            else:
                self._log("Could not get mesh vertices for color transfer", "WARNING")
        else:
            self._log("Source cloud has no colors (skipping transfer)")

        # Step 5: Save both cloud and mesh to single .bin file
        self._log_step(5, 5, "Saving project file...")

        # Ensure output directory exists
        output_file.parent.mkdir(parents=True, exist_ok=True)

        ret = cc.SaveEntities([cloud, mesh], str(output_file))
        if ret != 0:
            self._log(f"Failed to save: {output_file}", "ERROR")
            return False
        self._log(f"Saved: {output_file.name}", "SUCCESS")
        self._log(f"Successfully processed: {input_file.name}", "SUCCESS")
        return True

    def process_directory(
        self, input_dir: Path, output_subdir: str = "Processed"
    ) -> dict:
        """Process all LAS files in a directory."""
        input_dir = Path(input_dir).resolve()
        output_dir = input_dir / output_subdir

        self._log("=" * 70)
        self._log("CloudComPy Batch Processing")
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

        # Process files
        success_count = 0
        failed_count = 0

        for i, las_file in enumerate(las_files, 1):
            self._log(f"\nFile {i}/{len(las_files)}")
            output_file = output_dir / f"{las_file.stem}.bin"

            if self.process_file(las_file, output_file):
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
        self._log("  - [filename].bin : CloudCompare project with cloud and mesh")

        return {
            "total": len(las_files),
            "success": success_count,
            "failed": failed_count,
        }


def main():
    """Main entry point."""
    parser = argparse.ArgumentParser(
        description="Batch process LAS files with CloudComPy and PoissonRecon",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  %(prog)s /path/to/las/files
  %(prog)s . --output-dir Processed
  %(prog)s /data --octree-depth 12

Prerequisites:
  - CloudComPy: https://github.com/CloudCompare/CloudComPy

Output:
  Creates CloudCompare project files (.bin) containing:
  - Point cloud with normals and DIP/Dip Direction scalar fields
  - Reconstructed mesh with density scalar field

  The density SF allows filtering in CloudCompare:
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

    # Normal computation parameters
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
        "--boundary-type",
        type=int,
        choices=[0, 1, 2],
        default=2,
        help="Boundary type: 0=Free, 1=Dirichlet, 2=Neumann (default: 2)",
    )

    parser.add_argument(
        "--quiet",
        action="store_true",
        help="Suppress progress output",
    )

    args = parser.parse_args()

    # Create parameter objects
    normal_params = NormalParams(knn=args.knn)

    poisson_params = PoissonParams(
        octree_depth=args.octree_depth,
        samples_per_node=args.samples_per_node,
        point_weight=args.point_weight,
        boundary_type=args.boundary_type,
    )

    try:
        processor = CloudComPyProcessor(
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
