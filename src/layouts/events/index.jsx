/* eslint-disable tailwindcss/no-custom-classname */
import "./style.css";

import * as dayjs from "dayjs";

import ContentPanel, {
  ContentPanelBody,
  ContentPanelTitle,
  ContentPanelTitleIcon,
} from "../../components/content-panel";
import Pagination, { usePageHandler } from "../../components/pagination";
import { VscCloud, VscDebugStepInto, VscPlay } from "react-icons/vsc";

import { AutoSizer } from "react-virtualized";
import { Config } from "../../util";
import DirektivEditor from "../../components/editor";
import FlexBox from "../../components/flexbox";
import HelpIcon from "../../components/help";
import { Link } from "react-router-dom";
import Modal from "../../components/modal";
import relativeTime from "dayjs/plugin/relativeTime";
import { useApiKey } from "../../util/apiKeyProvider";
import { useEvents } from "../../hooks";
import { useState } from "react";
import utc from "dayjs/plugin/utc";

dayjs.extend(utc);
dayjs.extend(relativeTime);

function EventsPageWrapper(props) {
  const { namespace } = props;
  if (!namespace) {
    return null;
  }

  return <EventsPage namespace={namespace} />;
}

export default EventsPageWrapper;

const PAGE_SIZE = 8;

function EventsPage(props) {
  const { namespace } = props;
  const [apiKey] = useApiKey();

  // errHistory and errListeners TODO show error if one

  const historyPageHandler = usePageHandler(PAGE_SIZE);
  const listenersPageHandler = usePageHandler(PAGE_SIZE);

  const {
    eventHistory,
    eventListeners,
    eventListenersPageInfo,
    eventHistoryPageInfo,
    sendEvent,
    replayEvent,
  } = useEvents(Config.url, true, namespace, apiKey, {
    listeners: [listenersPageHandler.pageParams],
    history: [historyPageHandler.pageParams],
  });

  return (
    <>
      <FlexBox col gap style={{ paddingRight: "8px" }}>
        <FlexBox>
          <ContentPanel style={{ width: "100%" }}>
            <ContentPanelTitle>
              <ContentPanelTitleIcon>
                <VscCloud />
              </ContentPanelTitleIcon>
              <FlexBox style={{ display: "flex", alignItems: "center" }} gap>
                <div>Cloud Events History</div>
                <HelpIcon msg="A history of events that have hit this specific namespace." />
              </FlexBox>
              <SendEventModal sendEvent={sendEvent} />
            </ContentPanelTitle>
            <ContentPanelBody>
              <FlexBox col style={{ justifyContent: "space-between" }}>
                <div
                  style={{
                    maxHeight: "40vh",
                    overflowY: "auto",
                    fontSize: "12px",
                    minWidth: "100%",
                  }}
                >
                  <table
                    className="cloudevents-table"
                    style={{ minWidth: "440px", width: "100%" }}
                  >
                    <thead>
                      <tr>
                        <th>Type</th>
                        <th style={{ width: "250px" }}>Source</th>
                        <th>Time</th>
                        <th style={{ textAlign: "center" }}>Actions</th>
                      </tr>
                    </thead>
                    {eventHistory !== null &&
                    typeof eventHistory === typeof [] &&
                    eventHistory.length > 0 ? (
                      <tbody>
                        {eventHistory.map((obj) => (
                          <tr
                            key={obj.id}
                            style={{ borderBottom: "1px solid #f4f4f4" }}
                          >
                            <td
                              title={obj.type}
                              style={{
                                textOverflow: "ellipsis",
                                overflow: "hidden",
                              }}
                            >
                              {obj.type}
                            </td>
                            <td
                              title={obj.source}
                              style={{
                                textOverflow: "ellipsis",
                                overflow: "hidden",
                              }}
                            >
                              {obj.source}
                            </td>
                            <td>
                              {dayjs
                                .utc(obj.receivedAt)
                                .local()
                                .format("HH:mm:ss a")}
                            </td>
                            <td
                              style={{
                                textAlign: "center",
                                justifyContent: "center",
                              }}
                            >
                              <FlexBox className="gap center">
                                <div>
                                  <Modal
                                    className="run-workflow-modal"
                                    style={{ justifyContent: "flex-end" }}
                                    modalStyle={{
                                      color: "black",
                                      width: "360px",
                                    }}
                                    title="Retrigger Event"
                                    button={
                                      <>
                                        <VscPlay />{" "}
                                        <span className="hide-800">
                                          Retrigger
                                        </span>
                                      </>
                                    }
                                    buttonProps={{
                                      auto: true,
                                      color: "info",
                                    }}
                                    actionButtons={[
                                      {
                                        label: "Retrigger",

                                        onClick: async () => {
                                          await replayEvent(obj.id);
                                        },

                                        buttonProps: {
                                          variant: "contained",
                                          color: "primary",
                                        },
                                        closesModal: true,
                                      },
                                      {
                                        label: "Cancel",
                                        closesModal: true,
                                      },
                                    ]}
                                  >
                                    <FlexBox style={{ overflow: "hidden" }}>
                                      Are you sure you want to retrigger{" "}
                                      {obj.id}?
                                    </FlexBox>
                                  </Modal>
                                </div>
                                <div>
                                  <Modal
                                    className="run-workflow-modal"
                                    modalStyle={{
                                      color: "black",
                                      minWidth: "360px",
                                      width: "50vw",
                                      height: "40vh",
                                      minHeight: "680px",
                                    }}
                                    title="View Event"
                                    btnStyle={{ width: "unset" }}
                                    button={<span>View</span>}
                                    buttonProps={{
                                      auto: true,
                                      color: "info",
                                    }}
                                    actionButtons={[
                                      {
                                        label: "Close",
                                        closesModal: true,
                                      },
                                    ]}
                                  >
                                    <FlexBox col style={{ overflow: "hidden" }}>
                                      <AutoSizer>
                                        {({ height, width }) => (
                                          <DirektivEditor
                                            noBorderRadius
                                            value={atob(obj.cloudevent)}
                                            readonly={true}
                                            dlang="plaintext"
                                            options={{
                                              autoLayout: true,
                                            }}
                                            width={width}
                                            height={height}
                                          />
                                        )}
                                      </AutoSizer>
                                    </FlexBox>
                                  </Modal>
                                </div>
                              </FlexBox>
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    ) : (
                      <tbody>
                        <tr>
                          <td>
                            <FlexBox className="table-no-content">
                              No cloud events history
                            </FlexBox>
                          </td>
                        </tr>
                      </tbody>
                    )}
                  </table>
                </div>
                <FlexBox
                  row
                  style={{
                    justifyContent: "flex-end",
                    paddingBottom: "1em",
                    flexGrow: 0,
                  }}
                >
                  <Pagination
                    pageHandler={historyPageHandler}
                    pageInfo={eventHistoryPageInfo}
                  />
                </FlexBox>
              </FlexBox>
            </ContentPanelBody>
          </ContentPanel>
        </FlexBox>
        <FlexBox>
          <ContentPanel style={{ width: "100%" }}>
            <ContentPanelTitle>
              <ContentPanelTitleIcon>
                <VscDebugStepInto />
              </ContentPanelTitleIcon>
              <FlexBox style={{ display: "flex", alignItems: "center" }} gap>
                <div>Active Event Listeners</div>
                <HelpIcon msg="Current listeners in a namespace that are listening for a cloud a event." />
              </FlexBox>
            </ContentPanelTitle>
            <ContentPanelBody>
              <FlexBox col style={{ justifyContent: "space-between" }}>
                <div
                  style={{
                    maxHeight: "40vh",
                    overflowY: "auto",
                    fontSize: "12px",
                  }}
                >
                  <table
                    className="event-listeners-table"
                    style={{ width: "100%" }}
                  >
                    <thead>
                      <tr>
                        <th>Workflow</th>
                        <th>Type</th>
                        <th>Mode</th>
                        <th>Updated</th>
                        <th>Event Types</th>
                      </tr>
                    </thead>
                    {eventListeners !== null &&
                    typeof eventListeners === typeof [] &&
                    eventListeners?.length > 0 ? (
                      <tbody>
                        {eventListeners.map((obj, i) => (
                          <tr
                            key={i} // todo: get an id from the backend
                            style={{ borderBottom: "1px solid #f4f4f4" }}
                          >
                            <td
                              style={{
                                textOverflow: "ellipsis",
                                overflow: "hidden",
                              }}
                            >
                              <Link
                                style={{ color: "#2396d8" }}
                                to={`/n/${namespace}/explorer${obj.workflow}`}
                              >
                                {obj.workflow}
                              </Link>
                            </td>
                            <td
                              style={{
                                textOverflow: "ellipsis",
                                overflow: "hidden",
                              }}
                            >
                              {obj.instance !== "" ? (
                                <Link
                                  style={{ color: "#2396d8" }}
                                  to={`/n/${namespace}/instances/${obj.instance}`}
                                >
                                  {obj.instance.split("-")[0]}
                                </Link>
                              ) : (
                                "start"
                              )}
                            </td>
                            <td
                              style={{
                                textOverflow: "ellipsis",
                                overflow: "hidden",
                              }}
                            >
                              {obj.mode}
                            </td>
                            <td>
                              {dayjs
                                .utc(obj.updatedAt)
                                .local()
                                .format("HH:mm:ss a")}
                            </td>
                            <td className="event-split">
                              {obj.events.map((obj, i) => (
                                <span key={i}>{obj.type}</span>
                              ))}
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    ) : (
                      <tbody>
                        <tr>
                          <td>
                            <FlexBox className="table-no-content">
                              No active event listeners
                            </FlexBox>
                          </td>
                        </tr>
                      </tbody>
                    )}
                  </table>
                </div>
                <FlexBox
                  row
                  style={{
                    justifyContent: "flex-end",
                    paddingBottom: "1em",
                    flexGrow: 0,
                  }}
                >
                  <Pagination
                    pageHandler={listenersPageHandler}
                    pageInfo={eventListenersPageInfo}
                  />
                </FlexBox>
              </FlexBox>
            </ContentPanelBody>
          </ContentPanel>
        </FlexBox>
      </FlexBox>
    </>
  );
}

function SendEventModal(props) {
  const { sendEvent } = props;
  const [eventData, setEventData] = useState(`{
    "specversion" : "1.0",
    "type" : "com.github.pull.create",
    "source" : "https://github.com/cloudevents/spec/pull",
    "subject" : "123",
    "id" : "A234-1234-1234",
    "time" : "2018-04-05T17:31:00Z",
    "comexampleextension1" : "value",
    "comexampleothervalue" : 5,
    "datacontenttype" : "text/xml",
    "data" : "<much wow=\\"xml\\"/>"
}`);

  return (
    <>
      <Modal
        title="Send New Event"
        button={<span>Send New Event</span>}
        actionButtons={[
          {
            label: "Send",

            onClick: async () => {
              await sendEvent(eventData);
            },
            buttonProps: { variant: "contained", color: "primary" },
            closesModal: true,
          },
          {
            label: "Cancel",
            closesModal: true,
          },
        ]}
        noPadding
      >
        <FlexBox col gap style={{ overflow: "hidden" }}>
          <FlexBox style={{ minHeight: "40vh", minWidth: "70vw" }}>
            <AutoSizer>
              {({ height, width }) => (
                <DirektivEditor
                  noBorderRadius
                  value={eventData}
                  setDValue={setEventData}
                  dlang="json"
                  options={{
                    autoLayout: true,
                  }}
                  width={width}
                  height={height}
                />
              )}
            </AutoSizer>
          </FlexBox>
        </FlexBox>
      </Modal>
    </>
  );
}
