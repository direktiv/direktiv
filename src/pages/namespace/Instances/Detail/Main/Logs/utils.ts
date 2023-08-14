import moment from "moment";

export const formatTime = (time: string) => moment(time).format("HH:mm:ss.mm");
