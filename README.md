## Установка

1.  **Клонируйте репозиторий:**

    ```bash
    git clone https://github.com/Gprojects1/go-voting-bot
    ```

2.  **Заполните `.env` необходимыми переменными:**

    Отредактируйте `.env` файл и укажите значения для следующих переменных окружения:

    ```
    BOT_TOKEN=YOUR_BOT_TOKEN
    ```

3.  **Запустите приложение с помощью Docker Compose:**

    ```bash
    docker-compose up --build -d
    ```

5.  **Просмотр логов бота:**

    ```bash
    docker-compose logs -f app
    ```
