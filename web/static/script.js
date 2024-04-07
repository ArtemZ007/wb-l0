document.addEventListener('DOMContentLoaded', async function() {
    try {
        // Инициализация частиц на фоне страницы.
        await particlesJS.load('particles-js', 'particles.json');
        console.log('callback - particles.js config loaded');
    } catch (error) {
        console.error('Error loading particles.js config:', error);
    }

    // Получение формы по её идентификатору.
    const form = document.getElementById('uidForm');
    const submitButton = form.querySelector('button[type="submit"]');
    const originalButtonText = submitButton.textContent;

    // Обработчик события отправки формы.
    form.onsubmit = async function(e) {
        e.preventDefault(); // Предотвращение стандартного поведения формы.
        submitButton.textContent = 'Отправка...'; // Индикатор загрузки
        submitButton.disabled = true; // Отключение кнопки на время отправки

        // Создание объекта FormData из формы.
        const formData = new FormData(form);

        try {
            // Отправка данных формы на сервер методом POST.
            const response = await fetch('/', {
                method: 'POST',
                body: formData
            });

            // Обработка ответа от сервера.
            if (response.ok) {
                // Успешная отправка данных.
                alert('Value submitted successfully!');
            } else {
                // Ошибка при отправке данных.
                alert('Failed to submit value.');
            }
        } catch (error) {
            // Обработка ошибки при отправке данных.
            console.error('Error:', error);
            alert('An error occurred. Please try again.');
        } finally {
            submitButton.textContent = originalButtonText; // Восстановление текста кнопки
            submitButton.disabled = false; // Включение кнопки после отправки
        }
    };
});