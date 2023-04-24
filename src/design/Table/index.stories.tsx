import { Folder, MoreVertical } from "lucide-react";
import type { Meta, StoryObj } from "@storybook/react";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeaderCell,
  TableRow,
} from "./index";
import Button from "../Button";
import { Card } from "../Card";
import moment from "moment";

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
          {people.map((person, i) => (
            <TableRow key={i}>
              <TableCell>{person.name}</TableCell>
              <TableCell>{person.title}</TableCell>
              <TableCell>{person.email}</TableCell>
              <TableCell>{person.role}</TableCell>
              <TableCell className="flex items-center space-x-3">
                <Button variant="outline">Edit</Button>
                <Button variant="ghost" size="sm" icon>
                  <MoreVertical />
                </Button>
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

export const FileBrowser = () => (
  <Card className="flex flex-col space-y-5">
    <Table>
      <TableBody>
        {people.map((person, i) => (
          <TableRow key={i}>
            <TableCell className="flex space-x-3 hover:underline">
              <Folder className="h-5" />
              <a href="#" className="flex-1">
                {person.name}
              </a>
              <span className="text-gray-8 dark:text-gray-dark-8">
                {moment().fromNow()}
              </span>
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  </Card>
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
        {people.map((person, i) => (
          <TableRow key={i}>
            <TableCell>{person.name}</TableCell>
            <TableCell>{person.title}</TableCell>
            <TableCell>{person.email}</TableCell>
            <TableCell>{person.role}</TableCell>
            <TableCell className="flex items-center space-x-3">
              <Button variant="outline">Edit</Button>
              <Button variant="ghost" size="sm" icon>
                <MoreVertical />
              </Button>
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  </div>
);
