import Badge from "react-bootstrap/Badge";

import * as dayjs from "dayjs";
import relativeTime from "dayjs/plugin/relativeTime";

dayjs.extend(relativeTime);

export function TimeSince(date) {
    return dayjs(date).fromNow();
}

export function TimeSinceUnix(date) {
    return dayjs.unix(date).fromNow();
}

export const RemoteResourceState = Object.freeze({
    fetching: 0,
    successful: 1,
    failed: 2,
    notFound: 3,
});
export const ResourceRegex = new RegExp("^[a-z][a-z0-9._-]{1,34}[a-z0-9]$");

// This is a workaround, but might even be alright in handling most status
// TODO: Find a better way to handle all statuses
export function InstanceStatusParse(instanceStatus) {
    if (!instanceStatus) {
        return null;
    }

    let splitStatus = instanceStatus.split(":");

    return splitStatus[0];
}

export function InstanceBadge(instanceStatus) {
    if (!instanceStatus) {
        return <Badge variant="dark">Status: Not Set</Badge>;
    }

    switch (instanceStatus) {
        case "ended":
        case "complete":
            return <Badge variant="success">Status: Complete</Badge>;
        case "pending":
            return <Badge variant="warning">Status: Pending</Badge>;
        case "running":
            return <Badge variant="info">Status: Running</Badge>;
        case "failed":
            return <Badge variant="danger">Status: Failed</Badge>;
        case "listening":
            return <Badge variant="secondary">Status: Listening</Badge>;
        case "crashed":
            return <Badge variant="dark">Status: Crashed</Badge>;
        case "cancelled":
            return (
                <Badge style={{backgroundColor: "#777777"}} variant="primary">
                    Status: Cancelled
                </Badge>
            );
        default:
            return (
                <Badge variant="dark">Status: Unkown Type '${instanceStatus}'</Badge>
            );
    }
}
