#!/bin/bash
THREADS=1
INSTALL=1
CLEAN=0
HELP=0

while [[ $# -gt 0 ]]; do
    key="$1"

    case $key in
        --no-install)
            INSTALL=0

            shift
            ;;

        -c|--clean)
            CLEAN=1

            shift
            ;;

        -t|--threads)
            if [[ -n "$2" && "$2" =~ ^[0-9]+$ ]]; then
                THREADS="$2"
                shift 2
            else
                echo "Error: -t/--threads requires a positive integer argument"
                exit 1
            fi
            ;;
        
        -h|--help)
            HELP=1
            shift
            ;;
        
        *)
            shift
            ;;
    esac
done

if [ "$HELP" -gt 0 ]; then
    echo "Usage: install.sh [OPTIONS]"
    echo
    echo "Options:"
    echo "  --no-install      Build the tool without installing it to the system."
    echo "  -c --clean        Remove build files for the tool."
    echo "  -t --threads N    Specify the number of threads to use for building. 0 = CPU core count (default: 1)."
    echo "  -h --help         Display this help message."

    exit 0
fi

# We need to clean build files if clean is enabled.
if [ "$CLEAN" -gt 0 ]; then
    # Clean common files first.
    echo "Cleaning Common Repository build files..."
    make -C modules/common clean
    echo "Done."

    echo "Cleaning build files..."
    make clean
    echo "Done."

    exit 0
fi

# Set thread count to CPU core count if less than 1.
if [ "$THREADS" -lt 1]; then
    THREADS=$(nproc)
fi

echo "Building Packet Batch using $THREADS threads..."

# First, we want to build our common objects which includes LibYAML. Read the PB-Common directory for more information.

echo "Building JSON-C..."
make -j $THREADS jsonc
echo "Done..."

if [ "$INSTALL" -gt 0 ]; then
    echo "Installing JSON-C..."
    sudo make -j $THREADS jsonc_install
fi

# Next, build LibBPF objects.
echo "Building LibBPF..."
make -j $THREADS libbpf
echo "Done..."

# Now build our primary objects and executables.
echo "Building Main..."
make -j $THREADS
echo "Done..."

# Finally, install our binary. This must be ran by root or with sudo.
if [ "$INSTALL" -gt 0 ]; then
    echo "Installing Main..."
    sudo make -j $THREADS install
    echo "Done..."
fi

echo "Build and/or installation completed!"