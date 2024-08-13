file=$(dirname $0)/$1.test.sh

# Test if file exists
if [ ! -f $file ]; then
    echo "File $file does not exist"
    exit 1
fi

# Include the file
source $file

# Run the test
~/Downloads/bombardier.exe -c $concurrency \
                           -n $total_requests \
                           -m $method \
                           -H "Content-Type: application/json" \
                           -b $body \
                           $endpoint
