# Используйте официальный образ Go как базовый
FROM golang:alpine

# Установите рабочую директорию внутри контейнера
WORKDIR /app

# Скопируйте go.mod и go.sum в рабочую директорию
COPY go.mod ./
COPY go.sum ./

# Загрузите зависимости
RUN go mod download

# Скопируйте исходный код проекта
COPY . .

# Скомпилируйте приложение и поместите исполняемый файл в корень рабочей директории
RUN go build -o ../../app/cmd/server/main ../../app/cmd/server/main.go


# Укажите команду для запуска приложения
CMD [ "../../app/cmd/server/main" ]