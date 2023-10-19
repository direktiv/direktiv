import "../../AppLegacy.css";
import "./style.css";

import type { ComponentMeta, ComponentStory } from "@storybook/react";
import ContentPanel, {
  ContentPanelBody,
  ContentPanelFooter,
  ContentPanelTitle,
  ContentPanelTitleIcon,
} from "./index";

import FlexBox from "../flexbox";
import { VscCloud } from "react-icons/vsc";

export default {
  title: "Components/ContentPanel",
  component: ContentPanel,
} as ComponentMeta<typeof ContentPanel>;

const Template: ComponentStory<typeof ContentPanel> = (args) => {
  return (
    <FlexBox center>
      <ContentPanel {...args} style={{ height: "200px", width: "50%" }}>
        <ContentPanelTitle>
          <ContentPanelTitleIcon>
            <VscCloud />
          </ContentPanelTitleIcon>
          <FlexBox style={{ display: "flex", alignItems: "center" }} gap>
            <div>ContentPanelTitle</div>
          </FlexBox>
        </ContentPanelTitle>
        <ContentPanelBody>ContentPanelBody</ContentPanelBody>
        <ContentPanelFooter>ContentPanelFooter</ContentPanelFooter>
      </ContentPanel>
    </FlexBox>
  );
};

export const Primary = Template.bind({});
Primary.args = {
  grow: false,
};

Primary.story = {};
