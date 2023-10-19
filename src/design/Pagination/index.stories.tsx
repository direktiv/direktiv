import type { Meta, StoryObj } from "@storybook/react";
import { Pagination, PaginationLink } from "./index";
import { useState } from "react";

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
        icon="left"
        onClick={() => {
          if (page > 1) setPage(page - 1);
        }}
      />

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
        icon="right"
        onClick={() => {
          if (page < 3) setPage(page + 1);
        }}
      />
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

export const DefaultAlign = () => (
  <Pagination>
    <PaginationLink icon="left" />
    <PaginationLink>1</PaginationLink>
    <PaginationLink active>2</PaginationLink>
    <PaginationLink>3</PaginationLink>
    <PaginationLink icon="right" />
  </Pagination>
);

export const CenterPagination = () => (
  <Pagination align="center">
    <PaginationLink icon="left" />
    <PaginationLink>1</PaginationLink>
    <PaginationLink active>2</PaginationLink>
    <PaginationLink>3</PaginationLink>
    <PaginationLink icon="right" />
  </Pagination>
);

export const DisabledPaginationButtons = () => (
  <Pagination align="center">
    <PaginationLink icon="left" disabled />
    <PaginationLink>1</PaginationLink>
    <PaginationLink active>2</PaginationLink>
    <PaginationLink>3</PaginationLink>
    <PaginationLink icon="right" />
  </Pagination>
);

export const LeftPagination = () => (
  <Pagination align="left">
    <PaginationLink icon="left" />
    <PaginationLink>1</PaginationLink>
    <PaginationLink active>2</PaginationLink>
    <PaginationLink>3</PaginationLink>
    <PaginationLink icon="right" />
  </Pagination>
);

export const NoPageButtons = () => {
  const [page, setPage] = useState(100);
  return (
    <Pagination align="center">
      <PaginationLink
        icon="left"
        onClick={() => {
          if (page > 1) setPage(page - 1);
        }}
      />
      <PaginationLink>Page {page}</PaginationLink>
      <PaginationLink
        icon="right"
        onClick={() => {
          setPage(page + 1);
        }}
      />
    </Pagination>
  );
};
