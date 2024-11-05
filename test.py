import requests
from locust import HttpUser, task, between
from locust import web
from locust.main import main
class MyUser(HttpUser):
    wait_time = between(1, 1)  # Set wait time to 1 second

    @task
    def my_task(self):
        start_time = time.time()
        response = requests.get("drupal.testing.com.c52af81d372bb830.convox.cloud")
        end_time = time.time()
        elapsed_time = end_time - start_time
        print(f"Response time: {elapsed_time:.2f} seconds")

if __name__ == "__main__":
    main()
