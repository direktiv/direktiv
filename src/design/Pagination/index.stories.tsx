import type { Meta, StoryObj } from "@storybook/react";
import { Pagination, PaginationLink } from "./index";
import React, { useState } from "react";
import {
  RxChevronLeft,
  RxChevronRight,
  RxDoubleArrowLeft,
  RxDoubleArrowRight,
} from "react-icons/rx";

const meta = {
  title: "Components/Pagination",
  component: Pagination,
} satisfies Meta<typeof Pagination>;

export default meta;
type Story = StoryObj<typeof meta>;

const StoryComponent = ({ ...args }) => {
  const [page, setPage] = useState<number>(1);
  const handleClick = (index: number) => {
    setPage(index);
  };
  return (
    <Pagination {...args}>
      <PaginationLink
        icon
        onClick={() => {
          if (page > 1) setPage(page - 1);
        }}
      >
        <RxChevronLeft className="h-5 w-5" aria-hidden="true" />
      </PaginationLink>
      {new Array(3).fill(0).map((prop, key) => (
        <PaginationLink
          key={`pg-lint-${key}`}
          onClick={() => handleClick(key + 1)}
          active={page === key + 1}
        >
          {key + 1}
        </PaginationLink>
      ))}
      <PaginationLink
        icon
        onClick={() => {
          if (page < 3) setPage(page + 1);
        }}
      >
        <RxChevronRight className="h-5 w-5" aria-hidden="true" />
      </PaginationLink>
    </Pagination>
  );
};
export const Default: Story = {
  render: ({ ...args }) => <StoryComponent {...args} />,
  args: {
    children: "Pagination Content",
  },
  argTypes: {
    align: {
      description: "select variant",
      control: "select",
      options: ["left", "center", "right"],
      type: { name: "string", required: false },
    },
  },
};

export const NoPageButtons = () => {
  const [page, setPage] = useState(100);
  return (
    <Pagination align="center">
      <PaginationLink
        icon
        onClick={() => {
          if (page > 1) setPage(page - 1);
        }}
      >
        <RxDoubleArrowLeft className="h-5 w-5" aria-hidden="true" />
      </PaginationLink>
      <PaginationLink active>{`PAGE ${page}`}</PaginationLink>
      <PaginationLink
        icon
        onClick={() => {
          setPage(page + 1);
        }}
      >
        <RxDoubleArrowRight className="h-5 w-5" aria-hidden="true" />
      </PaginationLink>
    </Pagination>
  );
};
