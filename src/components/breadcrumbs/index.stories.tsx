import React from 'react';
import { ComponentStory, ComponentMeta } from '@storybook/react';
import { withRouter } from 'storybook-addon-react-router-v6';
import '../../App.css';
import './style.css'

import Breadcrumbs from './index';
import FlexBox from '../flexbox';

export default {
    title: 'Components/Breadcrumbs',
    component: Breadcrumbs,
    decorators: [withRouter],
} as ComponentMeta<typeof Breadcrumbs>;

const Template: ComponentStory<typeof Breadcrumbs> = (args) => {
    return (
        <FlexBox className="breadcrumbs-row">
            <Breadcrumbs {...args} />
        </FlexBox>
    )
};

export const Explorer = Template.bind({});
Explorer.args = {
    namespace: "direktiv"
};

Explorer.story = {
    parameters: {
        reactRouter: {
            routePath: '/n/direktiv/explorer/directory/workflow',
        }
    }
};

export const JQPlayground = Template.bind({});
JQPlayground.args = {
    namespace: "direktiv"
};
JQPlayground.story = {
    parameters: {
        reactRouter: {
            routePath: '/jq',
        },
        docs: {
            description: {
              story: 'Example breadcrumb on **JQ Playground** route.',
            },
          },
    }
};

export const AdditionalChildren = Template.bind({});
AdditionalChildren.args = {
    namespace: "direktiv",
    additionalChildren: (
    <div style={{display: "flex", justifyContent:"flex-end", alignItems:"center", marginLeft:"6px", width:"100%"}}>
        <div style={{backgroundColor: "#dde0e2", padding:"4px 12px", color:"#3f3f3f", borderRadius:"12px", cursor:"pointer", fontWeight:"bold"}}>
            ReadOnly
        </div>
    </div>
    )
};
AdditionalChildren.story = {
    parameters: {
        reactRouter: {
            routePath: '/n/direktiv/explorer/directory/workflow',
        },
        docs: {
            description: {
              story: 'Example breadcrumb with additional children. ReadOnly badge added in this example.',
            },
          },
    }
};