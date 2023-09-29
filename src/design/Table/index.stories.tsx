import { Contact, File, MoreVertical } from "lucide-react";
import type { Meta, StoryObj } from "@storybook/react";
import {
  NoPermissions as NoPermissionsComponent,
  NoResult,
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

const files = [
  { name: "app.js", date: "2023-02-05" },
  { name: "database.sql", date: "2023-06-20" },
  { name: "image.jpg", date: "2023-09-27" },
  { name: "index.html", date: "2023-01-01" },
  { name: "logo.svg", date: "2023-07-22" },
  { name: "package.json", date: "2023-05-17" },
  { name: "README.md", date: "2023-04-13" },
  { name: "report.pdf", date: "2023-08-25" },
  { name: "style.css", date: "2023-03-10" },
  { name: "video.mp4", date: "2023-10-30" },
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

export const TableAndCard = () => (
  <Card className="flex flex-col space-y-5">
    <Table>
      <TableBody>
        {files.map((file) => (
          <TableRow key={file.name}>
            <TableCell>
              <div className="flex space-x-3">
                <File className="h-5" />
                <a href="#" className="flex-1 hover:underline">
                  {file.name}
                </a>
                <span className="text-gray-8 dark:text-gray-dark-8">
                  {moment(file.date).fromNow()}
                </span>
              </div>
            </TableCell>
            <TableCell className="w-0">
              <Button
                variant="ghost"
                size="sm"
                onClick={(e) => e.preventDefault()}
                icon
              >
                <MoreVertical />
              </Button>
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

export const EmptyListWithHeader = () => (
  <Table>
    <TableHead>
      <TableRow>
        <TableHeaderCell>Name</TableHeaderCell>
        <TableHeaderCell>Title</TableHeaderCell>
        <TableHeaderCell>Email</TableHeaderCell>
        <TableHeaderCell>Role</TableHeaderCell>
      </TableRow>
    </TableHead>
    <TableBody>
      <TableCell colSpan={4}>
        <NoResult icon={Contact}>no data found</NoResult>
      </TableCell>
    </TableBody>
  </Table>
);

export const EmptyListWithoutHeader = () => (
  <Card>
    <NoResult icon={Contact}>no data found</NoResult>
  </Card>
);

export const EmptyListWithButton = () => (
  <Card>
    <NoResult icon={Contact} button={<Button>Some Button</Button>}>
      no data found
    </NoResult>
  </Card>
);

export const NoPermissions = () => (
  <Card>
    <NoPermissionsComponent>
      You do not have permission to view this page.
    </NoPermissionsComponent>
  </Card>
);
