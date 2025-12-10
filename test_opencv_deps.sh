#!/bin/bash
# Test nếu opencv-python-headless có cần libgl1 và libglib2.0-0 không

echo "Testing OpenCV dependencies..."

# Test without libs
docker run --rm ubuntu:22.04 bash -c "
apt-get update -qq && apt-get install -y -qq python3-pip > /dev/null 2>&1
pip3 install --quiet opencv-python-headless numpy pillow 2>&1 | grep -i 'error' || echo 'Install OK'
python3 -c 'import cv2; print(\"OpenCV version:\", cv2.__version__)' 2>&1
"

echo ""
echo "If you see 'cannot open shared object file' errors, we need those libs."
echo "If it works, we can remove libgl1 and libglib2.0-0!"
