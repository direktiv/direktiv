import { ComponentMeta, ComponentStory } from '@storybook/react';
import '../../App.css';
import './style.css';

import FlexBox from '../flexbox';
import {ModalHeadless} from './index';
import { useState } from 'react';
import Button from '../button';

export default {
  title: 'Components/ModalHeadless',
  component: ModalHeadless,
} as ComponentMeta<typeof ModalHeadless>;

const Template: ComponentStory<typeof ModalHeadless> = (args) => {
  const [showModal, setShowModal] = useState(false)
  return (
    <FlexBox col center style={{ height: "300px" }}>
      <Button onClick={()=>{setShowModal(true)}}>Open Modal</Button>
      <ModalHeadless {...args} visible={showModal} setVisible={setShowModal}>
        <FlexBox col center>
          <span>Body of Modal</span>
        </FlexBox>
      </ModalHeadless>
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
  withCloseButton: true
};