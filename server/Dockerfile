FROM python:3.9.7-slim

# Copy the contents of the server directory into the image
COPY . /server

# Set the working directory
WORKDIR /server

# Set the PYTHONPATH to include the current directory
ENV PYTHONPATH=/server

# Run the unit tests using unittest discovery
RUN python -m unittest discover -s tests -p "test_common.py"

# Set the entry point
ENTRYPOINT ["/bin/sh"]