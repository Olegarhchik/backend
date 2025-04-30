const statusBar = document.querySelector('.status-bar');
const notifications = Array.from(document.querySelectorAll('.notification'));

notifications.forEach((elem, index) => {
    setTimeout(() => {
        elem.classList.add('show');
    }, 250 * (index + 1));

    const copyButton = elem.querySelector('.notification-close');

    copyButton.addEventListener('click', () => {
        elem.remove();
    });
});

const copyButtons = Array.from(document.querySelectorAll('img.copy'));

function copyTextToClipboard(text) {
    const textarea = document.createElement('textarea');
    textarea.value = text;
    textarea.style.position = 'fixed';  
    textarea.style.left = '-9999px';

    document.body.appendChild(textarea);
    textarea.focus();
    textarea.select();

    document.execCommand('copy')
    document.body.removeChild(textarea);
}

copyButtons.forEach(button => {
    button.addEventListener('click', () => {
        text = button.previousElementSibling.children[1].textContent;
        copyTextToClipboard(text);

        button.setAttribute('src', 'img/done.png');

        setTimeout(() => {
            button.setAttribute('src', 'img/copy.png');
        }, 5000);
    });
});

const editButtons = Array.from(document.querySelectorAll('img.edit'));

editButtons.forEach(button => {
    button.addEventListener('click', () => {
        const inputs = button.parentElement.querySelectorAll(':disabled');
        
        inputs.forEach(input => {
            input.removeAttribute('disabled');
            button.remove();

            if (input.getAttribute('type') == 'radio')
                return;

            input.focus();
                
            if (input.getAttribute('type') == 'date')
                return;
            
            input.selectionEnd = input.value.length;
        });
    });
});