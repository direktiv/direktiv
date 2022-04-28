export const TextAreaWidgetYAML = (props) => {
    return (
        <textarea 
        type="text"
        class="field-description yaml"
        spellcheck="false" value={props.value}
        required={props.required}
        onChange={(event) => props.onChange(event.target.value)} 
        onKeyDown={(event) => {
            // 'event.key' will return the key as a string: 'Tab'
            // 'event.keyCode' will return the key code as a number: Tab = '9'
            // You can use either of them
            if (event.key  === 'Tab') {
                event.preventDefault();
                const { selectionStart, selectionEnd } = event.target;
                const oldValue = event.target.value
                const newValue = oldValue.substring(0, selectionStart) +"  " +oldValue.substring(selectionEnd)
                props.onChange(newValue)
            }
        }}
        />
    );
};

export const TextAreaWidgetJS = (props) => {
    return (
        <textarea 
        type="text"
        class="field-description yaml"
        spellcheck="false" value={props.value}
        required={props.required}
        onChange={(event) => props.onChange(event.target.value)} 
        onKeyDown={(event) => {
            // 'event.key' will return the key as a string: 'Tab'
            // 'event.keyCode' will return the key code as a number: Tab = '9'
            // You can use either of them
            if (event.key  === 'Tab') {
                event.preventDefault();
                const { selectionStart, selectionEnd } = event.target;
                const oldValue = event.target.value
                const newValue = oldValue.substring(0, selectionStart) +"  " +oldValue.substring(selectionEnd)
                props.onChange(newValue)
            }
        }}
        />
    );
};

export const CustomWidgets = {
    textAreaWidgetYAML: TextAreaWidgetYAML,
    textAreaWidgetJS: TextAreaWidgetJS
};