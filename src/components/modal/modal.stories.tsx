import { ComponentMeta, ComponentStory } from '@storybook/react';
import '../../App.css';
import './style.css';

import FlexBox from '../flexbox';
import Modal from './index';
import { useState } from 'react';

export default {
  title: 'Components/Modal',
  component: Modal,
} as ComponentMeta<typeof Modal>;

const TemplateState: ComponentStory<typeof Modal> = (args) => {
  const [counter, setCounter] = useState(0)
  return (
    <FlexBox col center style={{ height: "300px" }}>
      <Modal {...args}
        title="Counter"
        onClose={()=>{setCounter(0)}}
        requiredFields={[
          {tip: "counter must be at least 10", condition: counter >= 10},
        ]}
        actionButtons={[
          {
            label: "Increment Counter",
            buttonProps: {
              tooltip: "Increase counter by 1",
              variant: "contained",
              color: "primary"
            },
            onClick: () => {
              setCounter(counter + 1)
            },
          },
          {
            label: "Multiply Counter",
            buttonProps: {
              tooltip: "Multiply counter by 10",
              variant: "contained",
              color: "primary"
            },
            onClick: () => {
              setCounter(counter*10)
            },
            validate: true
          },
          {
            closesModal: true,
            label: "Cancel",
            buttonProps: {
              tooltip: "Closes Modal"
            }
          }
        ]}>
        <FlexBox col center>
          <span>Count: {counter}</span>
        </FlexBox>
      </Modal>
    </FlexBox>
  )
};

const Template: ComponentStory<typeof Modal> = (args) => {
  return (
    <FlexBox col center style={{ height: "300px" }}>
      <Modal {...args} >
        <FlexBox col center>
          <span>Body of Modal</span>
        </FlexBox>
      </Modal>
    </FlexBox>
  )
};

export const Simple = Template.bind({});
Simple.args = {
  actionButtons: [{
    closesModal: true,
    label: "Cancel",
    buttonProps: {
      tooltip: "Closes Modal"
    }
  }],
  modalStyle: {
    margin: "16px auto",
    maxWidth: "380px",
    maxHeight: "260px"
  },
  withCloseButton: true,
  label: "Open Modal"
};

export const Complex = TemplateState.bind({});
Complex.args = {
  modalStyle: {
    margin: "16px auto",
    maxWidth: "380px",
    maxHeight: "260px"
  },
  withCloseButton: true,
  button:
    <span>
      Open Modal
    </span>
};

Complex.story = {
  parameters: {
      docs: {
          description: {
            story: 'A more complex example of a Modal, that uses action buttons functions to increment a counter in the modals body. ' +
            'Also showcases the use of the validate and required fields props to show how to validate values on action buttons.',
          },
        },
  }
};