#!/bin/bash
JDK_URL="https://corretto.aws/downloads/resources/17.0.13.11.1/amazon-corretto-17.0.13.11.1-linux-x64.tar.gz"
JDK_DIR="$HOME/.jdks"
JDK_TAR="amazon-corretto-17.0.13.11.1-linux-x64.tar.gz"
DESIRED_DIR_NAME="amazon-corretto"

mkdir -p $JDK_DIR
if [ $? -ne 0 ]; then
    echo "Failed to create directory $JDK_DIR"
    exit 1
fi

cd $JDK_DIR
if [ $? -ne 0 ]; then
    echo "Failed to change directory to $JDK_DIR"
    exit 1
fi

echo "Downloading JDK from $JDK_URL..."
curl -L -o $JDK_TAR $JDK_URL
if [ $? -ne 0 ]; then
    echo "Failed to download JDK from $JDK_URL"
    exit 1
fi

echo "Extracting JDK..."
tar -xzf $JDK_TAR
if [ $? -ne 0 ]; then
    echo "Failed to extract $JDK_TAR"
    exit 1
fi

EXTRACTED_DIR=$(tar -tzf $JDK_TAR | head -1 | cut -f1 -d"/")
if [ -z "$EXTRACTED_DIR" ]; then
    echo "Failed to determine the extracted directory"
    exit 1
fi

echo "Renaming extracted directory to $DESIRED_DIR_NAME..."
mv "$JDK_DIR/$EXTRACTED_DIR" "$JDK_DIR/$DESIRED_DIR_NAME"
if [ $? -ne 0 ]; then
    echo "Failed to rename directory from $EXTRACTED_DIR to $DESIRED_DIR_NAME"
    exit 1
fi

rm $JDK_TAR
if [ $? -ne 0 ]; then
    echo "Failed to remove $JDK_TAR"
    exit 1
fi

echo "Setting JAVA_HOME and updating PATH..."
export JAVA_HOME="$JDK_DIR/$DESIRED_DIR_NAME"
export PATH="$PATH:$JAVA_HOME/bin"
if [ ! -d "$JAVA_HOME" ]; then
    echo "JAVA_HOME is set to an invalid directory: $JAVA_HOME"
    exit 1
fi

echo "Updating .bashrc for persistence..."
echo "export JAVA_HOME=$JAVA_HOME" >> ~/.bashrc
if [ $? -ne 0 ]; then
    echo "Failed to update .bashrc with JAVA_HOME"
    exit 1
fi

echo "export PATH=\$PATH:\$JAVA_HOME/bin" >> ~/.bashrc
if [ $? -ne 0 ]; then
    echo "Failed to update .bashrc with PATH"
    exit 1
fi

source ~/.bashrc
if [ $? -ne 0 ]; then
    echo "Failed to source .bashrc"
    exit 1
fi

echo "JDK installation and setup complete."
exit 0
