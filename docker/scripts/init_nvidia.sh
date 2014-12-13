
# are the nvidia devices already mounted?  
NUM_NVIDIA_DEVICES=$(ls /dev | grep -i nvidia | wc -l)
if (( $NUM_NVIDIA_DEVICES > 0 )); then
    # if so, we're done
    echo "Nvidia devices already mounted, nothing to do"
    exit 0
fi

echo "No nvidia devices found, continuing .."
lsmod | grep -i nvidia 

# are the nvidia kernel mods already installed?
NUM_NVIDIA_MODS=$(lsmod | grep -i nvidia | wc -l)
if (( $NUM_NVIDIA_MODS <= 0 )); then
    # download and install nvidia kernel mods
    echo "No nvidia kernel modules detected, installing"
    wget http://tleyden-misc.s3.amazonaws.com/elastic-thought/nvidia-kernel-modules/coreos_stable_494.4.0_hvm/kernelmods.tar.gz && tar xvfz kernelmods.tar.gz && /usr/sbin/insmod nvidia.ko && /usr/sbin/insmod nvidia-uvm.ko 
fi

echo "Nvidia kernel modules: "
lsmod | grep -i nvidia 

# mount nvidia devices

echo "Mounting nvidia devices"

# Count the number of NVIDIA controllers found.
NVDEVS=`lspci | grep -i NVIDIA`
N3D=`echo "$NVDEVS" | grep "3D controller" | wc -l`
NVGA=`echo "$NVDEVS" | grep "VGA compatible controller" | wc -l`
N=`expr $N3D + $NVGA - 1`
for i in `seq 0 $N`; do
mknod -m 666 /dev/nvidia$i c 195 $i
done
mknod -m 666 /dev/nvidiactl c 195 255

# Find out the major device number used by the nvidia-uvm driver
D=`grep nvidia-uvm /proc/devices | awk '{print $1}'`
mknod -m 666 /dev/nvidia-uvm c $D 0

echo "Done mounting nvidia devices"

ls /dev | grep -i nvidia 
