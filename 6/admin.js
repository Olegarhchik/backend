const userInputs = document.querySelectorAll('.users td:not(.non-editable) input');

userInputs.forEach(input => {
    input.addEventListener('click', () => {
        input.removeAttribute('readonly');
        input.parentElement.classList.remove('editable');
        input.removeEventListener('click', () => {});
    });
});

const declineButtons = document.querySelector('.decline');

declineButtons.addEventListener('click', (event) => {
    event.preventDefault();
    window.location.hash = '';
})

const removeButtons = document.querySelectorAll('.remove a');
let confirmInput = document.querySelector('#confirm p > input');

removeButtons.forEach(button => {
    button.addEventListener('click', () => {
        const login = button.parentElement
                            .parentElement
                            .querySelector('.login input')
                            .getAttribute('value');

        confirmInput.setAttribute('value', login);
    });
});