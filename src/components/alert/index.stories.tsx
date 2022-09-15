import React, { useState } from 'react';
import Alert, { AlertProps } from "./index"
import { ComponentStory, ComponentMeta } from '@storybook/react';
import '../../App.css';
import './style.css'
import Button from '../button';


export default {
    title: 'Components/Alert',
    component: Alert,
    argTypes: {
        severity: {
            control: { type: 'select' },
        },
        variant: {
            options: ['standard', 'filled', 'outlined'],
            control: { type: 'select' },
            defaultValue: 'standard',
        },
        grow: {
            control: { type: "boolean" },
        },
        children: {
            defaultValue: "Hello this is an example alert message!",
            control: { type: "text" }
        }
    },
} as ComponentMeta<typeof Alert>;

const TemplateState: ComponentStory<typeof Alert> = (args) => {
    const [errorMsg, setErrorMsg] = useState("fetching workflows" as null | string)
    return (
        <>
            {errorMsg ?
                <Alert {...args} onClose={() => { setErrorMsg(null) }}>
                    <span>Error: {errorMsg}</span>
                </Alert>
                :
                <Button color="error" onClick={() => { setErrorMsg("fetching workflows") }}>Show Alert</Button>
            }
        </>

    )
};
const Template: ComponentStory<typeof Alert> = (args) => <Alert {...args} />;


export const Default = Template.bind({});
Default.args = { severity: "info", grow: false };

export const Error = TemplateState.bind({});
Error.args = {
    severity: "error",
    grow: false,
};
Error.story = {
    parameters: {
        docs: {
            description: {
                story: 'An error alert example with the onClose prop being used to allow the user to hide the alert.'
            },
        },
    }
};