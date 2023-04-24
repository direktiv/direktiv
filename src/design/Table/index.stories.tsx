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
    name: "John Doe",
    title: "Back-end Developer",
    email: "john.doe@example.com",
    role: "Member",
  },
  {
    name: "Jane Smith",
    title: "UI/UX Designer",
    email: "jane.smith@example.com",
    role: "Member",
  },
  {
    name: "Michael Brown",
    title: "Project Manager",
    email: "michael.brown@example.com",
    role: "Member",
  },
  {
    name: "Sarah Johnson",
    title: "Data Analyst",
    email: "sarah.johnson@example.com",
    role: "Member",
  },
  {
    name: "Tom Wilson",
    title: "System Administrator",
    email: "tom.wilson@example.com",
    role: "Member",
  },
  {
    name: "Emily Lee",
    title: "Full-stack Developer",
    email: "emily.lee@example.com",
    role: "Member",
  },
  {
    name: "Kevin Chen",
    title: "Database Administrator",
    email: "kevin.chen@example.com",
    role: "Member",
  },
  {
    name: "David Rodriguez",
    title: "Software Engineer",
    email: "david.rodriguez@example.com",
    role: "Admin",
  },
  {
    name: "Karen Lee",
    title: "Front-end Developer",
    email: "karen.lee@example.com",
    role: "Admin",
  },
  {
    name: "James Johnson",
    title: "Back-end Developer",
    email: "james.johnson@example.com",
    role: "Member",
  },
  {
    name: "Christine Kim",
    title: "UI/UX Designer",
    email: "christine.kim@example.com",
    role: "Member",
  },
  {
    name: "Alex Nguyen",
    title: "Full-stack Developer",
    email: "alex.nguyen@example.com",
    role: "Member",
  },
  {
    name: "Amy Chen",
    title: "Project Manager",
    email: "amy.chen@example.com",
    role: "Member",
  },
  {
    name: "Brian Lee",
    title: "Software Engineer",
    email: "brian.lee@example.com",
    role: "Member",
  },
  {
    name: "Michelle Lee",
    title: "Data Analyst",
    email: "michelle.lee@example.com",
    role: "Admin",
  },
  {
    name: "Jasmine Kim",
    title: "Front-end Developer",
    email: "jasmine.kim@example.com",
    role: "Member",
  },
  {
    name: "Daniel Kim",
    title: "Back-end Developer",
    email: "daniel.kim@example.com",
    role: "Admin",
  },
  {
    name: "Ethan Chen",
    title: "Full-stack Developer",
    email: "ethan.chen@example.com",
    role: "Member",
  },
  {
    name: "Olivia Wang",
    title: "UI/UX Designer",
    email: "olivia.wang@example.com",
    role: "Admin",
  },
];

export const Default: Story = {
  render: ({ ...args }) => (
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
  ),
  tags: ["autodocs"],
  argTypes: {},
};

export const FileBrowser = () => (
  <Card className="flex flex-col space-y-5">
    <Table>
      <TableBody>
        {people.map((person) => (
          <TableRow key={person.email}>
            <TableCell className="flex space-x-3">
              <Folder className="h-5" />
              <a href="#" className="flex-1 hover:underline">
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
        {people.map((person) => (
          <TableRow key={person.email}>
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
