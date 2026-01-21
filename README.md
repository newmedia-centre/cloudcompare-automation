# CloudCompare + PoissonRecon Automation Script

Batch processing scripts for automating point cloud processing workflows using CloudCompare CLI and the standalone PoissonRecon tool. These scripts load LAS files, compute normals, perform Poisson Surface Reconstruction with density scalar field output, and save CloudCompare projects for manual mesh filtering.

## Features

- **Batch Processing**: Automatically process multiple LAS files in a directory
- **Normal Computation**: Calculate normals using Octree with MST orientation (optimized for LAS files)
- **Surface Reconstruction**: Poisson Surface Reconstruction via standalone PoissonRecon executable
- **Density Scalar Field**: Output density values for filtering low-confidence reconstruction areas
- **Manual Filtering Workflow**: CloudCompare projects ready for adjusting SF display and filtering by value
- **Cross-Platform**: Scripts available for Linux (Bash), Windows (Batch), and Python

## Workflow Overview

The scripts automate the processing pipeline, then you manually filter the results in CloudCompare:

### Automated Steps

1. **Import LAS File** → CloudCompare CLI
2. **Compute Normals** using Octree method (optimal for LAS point clouds)
   - Orientation: Minimum Spanning Tree (MST) with knn = 6
3. **Convert Normals to DIP** → Dip/Dip Direction format
4. **Export as PLY** with normals for PoissonRecon
5. **Poisson Surface Reconstruction** → PoissonRecon with density output
6. **Save CloudCompare Project** (.bin) containing point cloud and mesh

### Manual Filtering in CloudCompare

After processing, open the `.bin` file in CloudCompare to filter the mesh:

1. **Open** the `.bin` project file
2. **Select the mesh** in the DB Tree
3. **View Density SF**: Properties panel → Scalar Field → select "Density"
4. **Adjust SF Display Range**: 
   - In Properties panel, find "SF Display" section
   - Adjust the "displayed" **min value** to visualize low-density areas
   - Low-density areas (shown in blue/cool colors) are less reliable
5. **Filter by Value**: Edit → Scalar Fields → Filter by Value
   - Set minimum threshold to remove low-density vertices
   - This creates a new filtered mesh
6. **Export** the filtered mesh via File → Save

## Prerequisites

### CloudCompare
- Download and install from [cloudcompare.org](https://www.cloudcompare.org/)
- Ensure CLI support is available (included in standard installation)

### PoissonRecon
The standalone Poisson Surface Reconstruction tool from Michael Kazhdan:
- **Repository**: [github.com/mkazhdan/PoissonRecon](https://github.com/mkazhdan/PoissonRecon)
- **Pre-built Windows binaries**: Available in the repository releases
- **Linux/macOS**: Compile from source using the provided Makefile

#### Building PoissonRecon (Linux/macOS)
```bash
git clone https://github.com/mkazhdan/PoissonRecon.git
cd PoissonRecon
make
sudo cp Bin/Linux/PoissonRecon /usr/local/bin/
```

#### Windows Installation
1. Download pre-built executables from the [GitHub releases](https://github.com/mkazhdan/PoissonRecon/releases)
2. Extract to a folder (e.g., `C:\PoissonRecon\`)
3. Add to PATH or set `POISSONRECON_PATH` environment variable

## Installation

Clone or download this repository:

```bash
git clone https://github.com/yourusername/cloudcompare-automation-script.git
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

# Specify executable paths
python process_las_files.py /data \
    --cloudcompare-path /custom/path/CloudCompare \
    --poissonrecon-path /custom/path/PoissonRecon

# Or use environment variables
export CLOUDCOMPARE_PATH=/custom/path/CloudCompare
export POISSONRECON_PATH=/custom/path/PoissonRecon
python process_las_files.py /data
```

#### Python Script Options

| Option | Default | Description |
|--------|---------|-------------|
| `input_dir` | `.` | Directory containing LAS files |
| `--output-dir` | `Processed` | Subdirectory for output files |
| `--cloudcompare-path` | Auto-detect | Path to CloudCompare executable |
| `--poissonrecon-path` | Auto-detect | Path to PoissonRecon executable |
| `--radius` | `AUTO` | Radius for octree normal computation |
| `--knn` | `6` | K-nearest neighbors for MST orientation |
| `--octree-depth` | `11` | Octree depth for Poisson reconstruction |
| `--samples-per-node` | `1.5` | Samples per node for Poisson |
| `--point-weight` | `2.0` | Point weight for Poisson |
| `--threads` | `16` | Number of processing threads |
| `--boundary-type` | `3` | Boundary: 1=Free, 2=Dirichlet, 3=Neumann |
| `--no-colors` | `False` | Disable color interpolation |
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
:: Process LAS files in current directory
process_las_files.bat

:: Process LAS files in a specific directory
process_las_files.bat C:\path\to\las\files

:: Show help
process_las_files.bat /?
```

## Output

Processed files are saved in the `Processed/` subdirectory:

```
input_directory/
├── file1.las
├── file2.las
├── file3.las
└── Processed/
    ├── file1.bin    # CloudCompare project with point cloud + mesh
    ├── file2.bin
    └── file3.bin
```

### What's in the .bin file?

Each `.bin` CloudCompare project contains:
- **Point Cloud**: With computed normals and DIP (Dip/Dip Direction) values
- **Mesh**: Reconstructed surface with **Density** scalar field

The **Density** scalar field values indicate reconstruction confidence:
- **Higher values** = More sample points nearby = Higher confidence
- **Lower values** = Fewer sample points = Lower confidence (consider filtering)

## Processing Parameters

### Normal Computation (CloudCompare)

| Parameter | Value | Description |
|-----------|-------|-------------|
| Method | Octree | Optimal for large LAS point clouds |
| Radius | AUTO | Automatic radius estimation |
| Orientation | MST | Minimum Spanning Tree |
| KNN | 6 | Neighbors for MST orientation |

### Poisson Surface Reconstruction

| Parameter | Default | Description |
|-----------|---------|-------------|
| `--depth` | `11` | Maximum octree depth (higher = more detail, more memory) |
| `--samplesPerNode` | `1.5` | Minimum samples per node (1.0-5.0 for clean data, 15.0-20.0 for noisy) |
| `--pointWeight` | `2.0` | Interpolation weight (0 = original unscreened Poisson) |
| `--bType` | `3` | Boundary type: 1=Free, 2=Dirichlet, 3=Neumann |
| `--threads` | `16` | Number of processing threads |
| `--density` | Always ON | Output density SF for filtering |
| `--colors` | ON | Interpolate colors from input samples |

## Filtering Guide

### Understanding the Density Scalar Field

The density value at each mesh vertex indicates how well that area was sampled:
- Areas with **high density** (warm colors: red/orange) have good point coverage
- Areas with **low density** (cool colors: blue/purple) have sparse point coverage and may contain reconstruction artifacts

### Recommended Filtering Workflow

1. **Open the .bin file** in CloudCompare
2. **Select the mesh** in the DB Tree (left panel)
3. **Activate the Density SF**:
   - Properties panel (right side) → Scalar Fields
   - Select "Density" from the dropdown
4. **Examine the color distribution**:
   - Default coloring shows density variation
   - Identify the threshold where reliable data ends
5. **Adjust display range** to find your threshold:
   - Properties → SF Display → adjust "displayed" min value
   - Watch how the mesh coloring changes
6. **Filter the mesh**:
   - Edit → Scalar Fields → Filter by Value
   - Set your chosen minimum density threshold
   - Click "Split" or "Export" to create filtered mesh
7. **Save the result**:
   - Select the filtered mesh
   - File → Save → choose your preferred format (PLY, OBJ, etc.)

### Typical Density Thresholds

The optimal threshold depends on your data. Start with these guidelines:
- **Conservative** (keep more): Filter out bottom 5-10% of density values
- **Moderate**: Filter out areas below the median density
- **Aggressive** (cleaner result): Keep only top 50% density values

## Installation Paths

The scripts automatically search for executables in common locations:

### CloudCompare

| Platform | Paths |
|----------|-------|
| Linux | `/usr/bin/CloudCompare`, `/usr/bin/cloudcompare`, `/usr/local/bin/CloudCompare`, `/opt/CloudCompare/CloudCompare`, `/snap/bin/cloudcompare` |
| macOS | `/Applications/CloudCompare.app/Contents/MacOS/CloudCompare` |
| Windows | `C:\Program Files\CloudCompare\CloudCompare.exe`, `C:\Program Files (x86)\CloudCompare\CloudCompare.exe` |

### PoissonRecon

| Platform | Paths |
|----------|-------|
| Linux | `/usr/local/bin/PoissonRecon`, `/usr/bin/PoissonRecon`, `/opt/PoissonRecon/PoissonRecon` |
| macOS | `/usr/local/bin/PoissonRecon` |
| Windows | `C:\PoissonRecon\PoissonRecon.exe`, `C:\Program Files\PoissonRecon\PoissonRecon.exe` |

### Override with Environment Variables

```bash
# Linux/macOS
export CLOUDCOMPARE_PATH=/path/to/CloudCompare
export POISSONRECON_PATH=/path/to/PoissonRecon

# Windows (Command Prompt)
set CLOUDCOMPARE_PATH=C:\path\to\CloudCompare.exe
set POISSONRECON_PATH=C:\path\to\PoissonRecon.exe

# Windows (PowerShell)
$env:CLOUDCOMPARE_PATH = "C:\path\to\CloudCompare.exe"
$env:POISSONRECON_PATH = "C:\path\to\PoissonRecon.exe"
```

## Troubleshooting

### Executable Not Found

Ensure both CloudCompare and PoissonRecon are installed and accessible:

```bash
# Test CloudCompare
CloudCompare --version

# Test PoissonRecon
PoissonRecon --help
```

### Permission Denied (Linux/macOS)

Make the scripts executable:

```bash
chmod +x process_las_files.sh
chmod +x process_las_files.py
```

### PLY Export Issues

If CloudCompare fails to export PLY with normals:
1. Verify the input LAS file contains valid point data
2. Ensure sufficient memory is available for normal computation
3. Check the output directory is writable

### PoissonRecon Memory Issues

For very large point clouds:
1. Reduce `--depth` parameter (e.g., from 11 to 9 or 10)
2. Use `--maxMemory` flag to limit memory usage
3. Process files sequentially
4. Ensure sufficient RAM is available (rule of thumb: 4GB+ for depth 10, 8GB+ for depth 11)

### Missing Density Scalar Field

If the mesh doesn't show the Density SF:
1. Verify PoissonRecon ran with `--density` flag (always enabled in these scripts)
2. In CloudCompare, ensure Scalar Fields is enabled in Properties panel
3. Select "Density" from the SF dropdown list

### Mesh Has Holes or Artifacts

Try adjusting PoissonRecon parameters:
- **Increase** `--depth` for more detail (but more memory/time)
- **Decrease** `--samplesPerNode` for denser reconstruction
- **Increase** `--pointWeight` for tighter fit to input points

## Contributing

Contributions are welcome! Please feel free to submit issues or pull requests.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [CloudCompare](https://www.cloudcompare.org/) - Open source 3D point cloud and mesh processing software
- [PoissonRecon](https://github.com/mkazhdan/PoissonRecon) - Poisson Surface Reconstruction by Michael Kazhdan
- Based on the papers by Kazhdan, Bolitho, Hoppe (2006) and Kazhdan, Hoppe (2013)