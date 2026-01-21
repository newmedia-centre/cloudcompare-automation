# CloudCompare Automation Script

Batch processing scripts for CloudCompare to automate point cloud processing workflows. These scripts load LAS files, compute normals, perform Poisson Surface Reconstruction, and save projects for manual refinement.

## Features

- **Batch Processing**: Automatically process multiple LAS files in a directory
- **Normal Computation**: Calculate normals using Triangulation with MST orientation
- **Surface Reconstruction**: Poisson Surface Reconstruction with configurable parameters
- **Cross-Platform**: Scripts available for Linux (Bash), Windows (Batch), and Python
- **Configurable**: All processing parameters can be customized

## Processing Pipeline

For each LAS file, the scripts perform the following operations:

1. **Import LAS File** - Load the point cloud data
2. **Compute Normals**
   - Local surface model: Triangulation
   - Orientation: Minimum Spanning Tree (MST) with knn = 6
3. **Convert Normals to DIP** - Transform normals to Dip/Dip Direction format
4. **Poisson Surface Reconstruction**
   - Octree depth: 11
   - Samples per node: 1.5
   - Point weight: 2.0
   - Threads: 16
   - Boundary: Neumann
   - Output density as SF: Yes
   - Interpolate cloud colors: Yes
5. **Save Project** - Export as `[filename].bin` in the `Processed/` folder

## Installation

### Prerequisites

- [CloudCompare](https://www.cloudcompare.org/) installed on your system
- Python 3.7+ (for the Python script)

### Download

Clone or download this repository:

```bash
git clone https://github.com/newmedia-centre/cloudcompare-automation-script.git
cd cloudcompare-automation-script
```

## Usage

### Python Script (Recommended)

The Python script provides the most flexibility and cross-platform support:

```bash
# Process LAS files in current directory
python process_las_files.py

# Process LAS files in a specific directory
python process_las_files.py /path/to/las/files

# Customize parameters
python process_las_files.py /data --octree-depth 12 --threads 8 --knn 10

# Specify CloudCompare path
python process_las_files.py /data --cloudcompare-path /custom/path/CloudCompare

# Or use environment variable
export CLOUDCOMPARE_PATH=/custom/path/CloudCompare
python process_las_files.py /data
```

#### Python Script Options

| Option | Default | Description |
|--------|---------|-------------|
| `input_dir` | `.` | Directory containing LAS files |
| `--output-dir` | `Processed` | Subdirectory for output files |
| `--cloudcompare-path` | Auto-detect | Path to CloudCompare executable |
| `--octree-depth` | `11` | Octree depth for Poisson reconstruction |
| `--samples-per-node` | `1.5` | Samples per node for Poisson |
| `--point-weight` | `2.0` | Point weight for Poisson |
| `--threads` | `16` | Number of processing threads |
| `--boundary` | `NEUMANN` | Boundary condition (FREE, DIRICHLET, NEUMANN) |
| `--knn` | `6` | K-nearest neighbors for MST orientation |
| `--quiet` | `False` | Suppress progress output |

### Bash Script (Linux/macOS)

```bash
# Make the script executable
chmod +x process_las_files.sh

# Process LAS files in current directory
./process_las_files.sh

# Process LAS files in a specific directory
./process_las_files.sh /path/to/las/files

# Show help
./process_las_files.sh --help
```

### Batch Script (Windows)

```batch
# Process LAS files in current directory
process_las_files.bat

# Process LAS files in a specific directory
process_las_files.bat C:\path\to\las\files

# Show help
process_las_files.bat /?
```

## Output

Processed files are saved in the `Processed/` subdirectory (relative to the input directory):

```
input_directory/
├── file1.las
├── file2.las
├── file3.las
└── Processed/
    ├── file1.bin
    ├── file1_mesh.bin
    ├── file2.bin
    ├── file2_mesh.bin
    ├── file3.bin
    └── file3_mesh.bin
```

The `.bin` files are CloudCompare native project files that contain:
- The processed point cloud with computed normals and DIP values
- The reconstructed mesh from Poisson Surface Reconstruction

## CloudCompare CLI Reference

### Command Structure

The scripts use CloudCompare's command-line interface with the following structure:

```bash
CloudCompare -SILENT -AUTO_SAVE OFF \
    -O input.las \
    -COMPUTE_NORMALS \
    -ORIENT_NORMS_MST 6 \
    -NORMALS_TO_DIP \
    -POISSON 11 1.5 NEUMANN \
    -C_EXPORT_FMT BIN \
    -M_EXPORT_FMT BIN \
    -SAVE_CLOUDS FILE output.bin \
    -SAVE_MESHES FILE output_mesh.bin
```

### CLI Limitations

Some advanced Poisson parameters may not be available in the CLI and might require GUI-based processing:

- `point_weight` - May not be available in all CLI versions
- `threads` - Thread count configuration
- `output_density_as_sf` - Density scalar field output
- `interpolate_colors` - Color interpolation

For full parameter control, consider using CloudCompare's GUI batch processing feature or the Python bindings.

## CloudCompare Installation Paths

The scripts automatically search for CloudCompare in common installation locations:

### Linux
- `/usr/bin/CloudCompare`
- `/usr/bin/cloudcompare`
- `/usr/local/bin/CloudCompare`
- `/opt/CloudCompare/CloudCompare`
- `/snap/bin/cloudcompare`

### macOS
- `/Applications/CloudCompare.app/Contents/MacOS/CloudCompare`

### Windows
- `C:\Program Files\CloudCompare\CloudCompare.exe`
- `C:\Program Files (x86)\CloudCompare\CloudCompare.exe`

You can override these by setting the `CLOUDCOMPARE_PATH` environment variable.

## Troubleshooting

### CloudCompare Not Found

If the scripts can't find CloudCompare:

```bash
# Linux/macOS
export CLOUDCOMPARE_PATH=/path/to/CloudCompare

# Windows (Command Prompt)
set CLOUDCOMPARE_PATH=C:\path\to\CloudCompare.exe

# Windows (PowerShell)
$env:CLOUDCOMPARE_PATH = "C:\path\to\CloudCompare.exe"
```

### Permission Denied (Linux/macOS)

Make the script executable:

```bash
chmod +x process_las_files.sh
chmod +x process_las_files.py
```

### CLI Command Errors

Different versions of CloudCompare may have slightly different CLI syntax. If you encounter errors:

1. Check your CloudCompare version: `CloudCompare --version`
2. Consult the [CloudCompare Wiki](https://www.cloudcompare.org/doc/wiki/index.php/Command_line_mode) for your version's CLI documentation
3. Modify the script parameters as needed

### Large File Processing

For very large point clouds:

1. Increase the timeout in the scripts
2. Reduce octree depth to speed up processing
3. Process files sequentially rather than in parallel
4. Ensure sufficient RAM is available

## Contributing

Contributions are welcome! Please feel free to submit issues or pull requests.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [CloudCompare](https://www.cloudcompare.org/) - Open source 3D point cloud and mesh processing software
- The CloudCompare development team for providing CLI functionality
