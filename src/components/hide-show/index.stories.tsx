import "../../AppLegacy.css";
import "./style.css";

import type { ComponentMeta, ComponentStory } from "@storybook/react";

import HideShowButton from "./index";

export default {
  title: "Components/HideShowButton",
  component: HideShowButton,
} as ComponentMeta<typeof HideShowButton>;

const Template: ComponentStory<typeof HideShowButton> = (args) => {
  return <HideShowButton {...args} />;
};

export const Password = Template.bind({});
Password.args = {
  show: false,
  field: "password",
};
