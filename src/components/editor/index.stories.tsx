import { ComponentMeta, ComponentStory } from '@storybook/react';
import { AutoSizer } from 'react-virtualized';
import '../../App.css';
import './style.css';
import FlexBox from '../flexbox';

import DirektivEditor from './index';

const exampleWorkflow = `description: A simple 'no-op' state that returns 'Hello world!'
states:
- id: helloworld
  type: noop
  transition: exit
  transform:
    result: Hello world!
- id: exit
  type: noop
`

const exampleJSON = `{
    "result": "Hello world!"
}`

export default {
    title: 'Components/DirektivEditor',
    component: DirektivEditor,
} as ComponentMeta<typeof DirektivEditor>;

const Template: ComponentStory<typeof DirektivEditor> = (args) => {
    return (
        <FlexBox col row style={{ minWidth: "460px", width: "460px", height: "380px", minHeight: "380px" }}>
            <FlexBox style={{ overflow: "hidden" }}>
                {/* <AutoSizer> */}
                    {/* {({ height, width }) => ( */}
                        <DirektivEditor {...args}
                            width={460}
                            height={380}
                        />
                    {/* )} */}
                {/* </AutoSizer> */}
            </FlexBox>
        </FlexBox>

    )
};

export const YAMLExample = Template.bind({});
YAMLExample.args = {
    value: exampleWorkflow,
    dlang: "yaml"
};

export const ReadOnlyExample = Template.bind({});
ReadOnlyExample.args = {
    value: exampleJSON,
    dlang: "json",
    readonly: true
};
ReadOnlyExample.story = {
    parameters: {
        docs: {
            description: {
              story: 'Editor in Read Only mode with language (dlang) set to "json"',
            },
           },
    }
};

export const NoDataExample = Template.bind({});
NoDataExample.args = {
    dvalue: "No value",
};

NoDataExample.story = {
    parameters: {
        docs: {
            description: {
              story: 'Example of showing dvalue being display when value is not set.',
            },
          },
    }
};