document.addEventListener('DOMContentLoaded', async function () {
    // Инициализация particles.js с учетом кэширова��ия.
    try {
        await particlesJS.load('particles-js', 'static/js/particles.json?cacheBuster=' + new Date().getTime(), function () {
            console.log('callback - Конфигурация particles.js успешно загружена.');
        });
    } catch (error) {
        console.error('Ошибка при загрузке конфигурации particles.js:', error);
    }

    const form = document.getElementById('uidForm');
    const resultContainer = document.getElementById('result');
    const submitButton = form.querySelector('button[type="submit"]');
    const originalButtonText = submitButton.textContent;

    function validateFormData(formData) {
        const hashUID = formData.get('hash_uid').trim();
        if (!hashUID) {
            alert('Пожалуйста, заполните все необходимые поля.');
            return false;
        }
        return true;
    }

    form.addEventListener('submit', async function (e) {
        e.preventDefault();

        // Анимация кнопки
        submitButton.innerHTML = '<span class="spinner-border spinner-border-sm" role="status" aria-hidden="true"></span> Отправка...';
        submitButton.disabled = true;

        const formData = new FormData(form);

        if (!validateFormData(formData)) {
            submitButton.innerHTML = originalButtonText;
            submitButton.disabled = false;
            return;
        }

        const hashUID = formData.get('hash_uid').trim();

        try {
            const response = await fetch(`/api/orders/${hashUID}`, {
                method: 'GET',
                headers: {
                    'Cache-Control': 'no-cache'
                }
            });

            if (response.ok) {
                const data = await response.json();
                resultContainer.innerHTML = `<pre>${JSON.stringify(data, null, 2)}</pre>`;
                $(resultContainer).fadeIn().delay(5000).fadeOut();
            } else if (response.status === 404) {
                resultContainer.textContent = "Неверный UID";
                $(resultContainer).fadeIn().delay(5000).fadeOut();
            } else {
                throw new Error('Произошла ошибка при запросе');
            }
        } catch (error) {
            console.error('Ошибка:', error);
            resultContainer.textContent = "Произошла ошибка при запросе";
            $(resultContainer).fadeIn().delay(5000).fadeOut();
        } finally {
            submitButton.innerHTML = originalButtonText;
            submitButton.disabled = false;
        }
    });
});
