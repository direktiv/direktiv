import { withRouter } from 'storybook-addon-react-router-v6';

import '../../App.css';
import FlexBox from '../flexbox';
import './style.css';

import NamespaceSelector from './index';

export default {
    title: 'Components/NamespaceSelector',
    component: NamespaceSelector,
    decorators: [withRouter],
    argTypes: {
        namespace: {
            description: "Currently selected namespace. This value is handled by the parent and based on the current route.",
            table: {
                type: { summary: 'string' },
            }
        },
        namespaces: {
            description: "Array of available namespaces.",
        },
    }
};

const Template = (args) => {
    return (
        <FlexBox col center style={{ minHeight: "380px"}}>
            <FlexBox col center style={{ width: "250px"}}>
                <NamespaceSelector {...args}/>
            </FlexBox>
        </FlexBox>
    )
};

export const SelectNamespace = Template.bind({});
SelectNamespace.args = {
    namespaces: [
        {name: "direktiv"},
        {name: "prod"},
        {name: "dev"}
    ],
    namespace: "direktiv",
};