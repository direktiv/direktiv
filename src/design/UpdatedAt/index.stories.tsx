import type { Meta } from "@storybook/react";
import UpdatedAt from "./index";
import moment from "moment";

const meta = {
  title: "Components/UpdateAt",
  component: UpdatedAt,
} satisfies Meta<typeof UpdatedAt>;

export default meta;

export const Default = () => <UpdatedAt date={new Date()} />;

export const UpdatesForOneHour = () => <UpdatedAt date={moment()} />;
export const WillNotUpdate = () => <UpdatedAt date={moment("12.20.2022")} />;
