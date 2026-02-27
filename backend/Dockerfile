
FROM public.ecr.aws/docker/library/python:3.10-slim

WORKDIR /app

# Install system dependencies (for playwright/lxml if needed)
RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    && rm -rf /var/lib/apt/lists/*

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# If using Playwright
# RUN playwright install chromium
# RUN playwright install-deps

COPY . .

CMD ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "8000"]
