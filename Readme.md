Check for site availability via prometheus and grafana

Prometheus job config
````
- job_name: 'web'
  scrape_interval: 5s
  scrape_timeout: 5s
  metrics_path: /metrics
  scheme: http
  static_configs:
    - targets: ['web-exporter:5555']
````

Usage:
````
  docker-compose up -d
````
[MIT License](https://choosealicense.com/licenses/mit/)
