FROM python:3.12-slim

WORKDIR /app

# Install system dependencies if needed (e.g. for mysqlclient if used, but we'll use connector-python which is pure python mostly)
# RUN apt-get update && apt-get install -y default-libmysqlclient-dev build-essential

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY . .

# Keep container running to allow exec commands, or define a default command.
# Since we run scripts manually or via cron typically, "tail -f /dev/null" allows us to exec into it.
# Or we can make it run the batch immediately.
# Let's make it wait so we can attach and run commands as requested by user prompts usually.
CMD ["tail", "-f", "/dev/null"]
