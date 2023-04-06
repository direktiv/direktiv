import type { Meta, StoryObj } from "@storybook/react";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeaderCell,
  TableRow,
} from "./index";

const meta = {
  title: "Components/Table",
  component: Table,
} satisfies Meta<typeof Table>;

export default meta;
type Story = StoryObj<typeof meta>;

const people = [
  {
    name: "Lindsay Walton",
    title: "Front-end Developer",
    email: "lindsay.walton@example.com",
    role: "Member",
  },
  {
    name: "Lindsay Walton",
    title: "Front-end Developer",
    email: "lindsay.walton@example.com",
    role: "Member",
  },
  {
    name: "Lindsay Walton",
    title: "Front-end Developer",
    email: "lindsay.walton@example.com",
    role: "Member",
  },
  {
    name: "Lindsay Walton",
    title: "Front-end Developer",
    email: "lindsay.walton@example.com",
    role: "Member",
  },
  {
    name: "Lindsay Walton",
    title: "Front-end Developer",
    email: "lindsay.walton@example.com",
    role: "Member",
  },
  {
    name: "Lindsay Walton",
    title: "Front-end Developer",
    email: "lindsay.walton@example.com",
    role: "Member",
  },
  {
    name: "Lindsay Walton",
    title: "Front-end Developer",
    email: "lindsay.walton@example.com",
    role: "Member",
  },
  {
    name: "Lindsay Walton",
    title: "Front-end Developer",
    email: "lindsay.walton@example.com",
    role: "Member",
  },
  {
    name: "Lindsay Walton",
    title: "Front-end Developer",
    email: "lindsay.walton@example.com",
    role: "Member",
  },
  {
    name: "Lindsay Walton",
    title: "Front-end Developer",
    email: "lindsay.walton@example.com",
    role: "Member",
  },
  {
    name: "Lindsay Walton",
    title: "Front-end Developer",
    email: "lindsay.walton@example.com",
    role: "Member",
  },
  {
    name: "Lindsay Walton",
    title: "Front-end Developer",
    email: "lindsay.walton@example.com",
    role: "Member",
  },
  {
    name: "Lindsay Walton",
    title: "Front-end Developer",
    email: "lindsay.walton@example.com",
    role: "Member",
  },
  {
    name: "Lindsay Walton",
    title: "Front-end Developer",
    email: "lindsay.walton@example.com",
    role: "Member",
  },
  {
    name: "Lindsay Walton",
    title: "Front-end Developer",
    email: "lindsay.walton@example.com",
    role: "Member",
  },
  {
    name: "Lindsay Walton",
    title: "Front-end Developer",
    email: "lindsay.walton@example.com",
    role: "Member",
  },
  {
    name: "Lindsay Walton",
    title: "Front-end Developer",
    email: "lindsay.walton@example.com",
    role: "Member",
  },
  {
    name: "Lindsay Walton",
    title: "Front-end Developer",
    email: "lindsay.walton@example.com",
    role: "Member",
  },
  {
    name: "Lindsay Walton",
    title: "Front-end Developer",
    email: "lindsay.walton@example.com",
    role: "Member",
  },
  {
    name: "Lindsay Walton",
    title: "Front-end Developer",
    email: "lindsay.walton@example.com",
    role: "Member",
  },
  // More people...
];

export const Default: Story = {
  render: ({ ...args }) => (
    <>
      <Table {...args}>
        <TableHead>
          <TableRow>
            <TableHeaderCell>Name</TableHeaderCell>
            <TableHeaderCell>Title</TableHeaderCell>
            <TableHeaderCell>Email</TableHeaderCell>
            <TableHeaderCell>Role</TableHeaderCell>
            <TableHeaderCell>
              <span className="sr-only">Edit</span>
            </TableHeaderCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {people.map((person) => (
            <TableRow key={person.email}>
              <TableCell className="whitespace-nowrap py-4 pl-4 pr-3 text-sm font-medium text-gray-12 ">
                {person.name}
              </TableCell>
              <TableCell>{person.title}</TableCell>
              <TableCell>{person.email}</TableCell>
              <TableCell>{person.role}</TableCell>
              <TableCell className="sm:pr-0">
                <a className="text-primary-600 hover:text-primary-700">
                  Edit<span className="sr-only">, {person.name}</span>
                </a>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </>
  ),
  tags: ["autodocs"],
  argTypes: {},
};

export const StripeTable = () => (
  <>
    <Table>
      <TableHead>
        <TableRow>
          <TableHeaderCell>Name</TableHeaderCell>
          <TableHeaderCell>Title</TableHeaderCell>
          <TableHeaderCell>Email</TableHeaderCell>
          <TableHeaderCell>Role</TableHeaderCell>
          <TableHeaderCell className="relative py-3.5 pl-3 pr-4">
            <span className="sr-only">Edit</span>
          </TableHeaderCell>
        </TableRow>
      </TableHead>
      <TableBody>
        {people.map((person, key) => (
          <TableRow key={person.email} stripe={key % 2 === 0}>
            <TableCell className="font-medium text-gray-12">
              {person.name}
            </TableCell>
            <TableCell>{person.title}</TableCell>
            <TableCell>{person.email}</TableCell>
            <TableCell>{person.role}</TableCell>
            <TableCell className="font-medium sm:pr-0">
              <a className="text-primary-600 hover:text-primary-900">
                Edit<span className="sr-only">, {person.name}</span>
              </a>
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  </>
);

export const StickyHeader = () => (
  <div className="h-64">
    <Table>
      <TableHead>
        <TableRow>
          <TableHeaderCell sticky>Name</TableHeaderCell>
          <TableHeaderCell sticky>Title</TableHeaderCell>
          <TableHeaderCell sticky>Email</TableHeaderCell>
          <TableHeaderCell sticky>Role</TableHeaderCell>
          <TableHeaderCell sticky>
            <span className="sr-only">Edit</span>
          </TableHeaderCell>
        </TableRow>
      </TableHead>
      <TableBody>
        {people.map((person) => (
          <TableRow key={person.email}>
            <TableCell className="whitespace-nowrap py-4 pl-4 pr-3 text-sm font-medium text-gray-12 ">
              {person.name}
            </TableCell>
            <TableCell>{person.title}</TableCell>
            <TableCell>{person.email}</TableCell>
            <TableCell>{person.role}</TableCell>
            <TableCell className="sm:pr-0">
              <a className="text-primary-600 hover:text-primary-700">
                Edit<span className="sr-only">, {person.name}</span>
              </a>
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  </div>
);
