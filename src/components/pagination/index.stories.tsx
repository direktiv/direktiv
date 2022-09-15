import { ComponentMeta, ComponentStory } from '@storybook/react';
import '../../App.css';

import Pagination, { PageInfo, usePageHandler } from './index';

const examplePageInfo: PageInfo = {
    "order": [
      {
        "field": "NAME",
        "direction": ""
      }
    ],
    "filter": [
      {
        "field": "NAME",
        "type": "CONTAINS",
        "val": ""
      }
    ],
    "limit": 10,
    "offset": 0,
    "total": 50
} 

const PAGE_SIZE = 10

export default {
  title: 'Components/Pagination',
  component: Pagination,
} as ComponentMeta<typeof Pagination>;

const Template: ComponentStory<typeof Pagination> = (args) => {
  const pageHandler = usePageHandler(PAGE_SIZE)
  return (<Pagination {...args} pageHandler={pageHandler}/>)
};

export const Primary = Template.bind({});
Primary.args = {
  pageInfo: examplePageInfo
};