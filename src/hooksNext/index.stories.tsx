import moment from "moment";
import useUpdatedAt from "./useUpdatedAt";

const meta = {
  title: "Components/Hooks/useUpdatedAt",
};

export default meta;

export const Default = () => {
  const updatedAt = useUpdatedAt(new Date());
  return updatedAt;
};

export const UpdatesForOneHour = () => {
  const updatedAt = useUpdatedAt(moment());
  return updatedAt;
};

export const WillNotUpdate = () => {
  const updatedAt = useUpdatedAt(moment("12.20.2022"));
  return updatedAt;
};
