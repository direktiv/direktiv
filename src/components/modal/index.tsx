import "./style.css";
import "tippy.js/dist/tippy.css";

import Button, { ButtonProps } from "../button";
import ContentPanel, {
  ContentPanelBody,
  ContentPanelFooter,
  ContentPanelTitle,
  ContentPanelTitleIcon,
} from "../content-panel";
import { VscClose, VscDiffAdded } from "react-icons/vsc";
import { useEffect, useState } from "react";

import Alert from "../alert";
import FlexBox from "../flexbox";

export interface RequiredField {
  /**
   * Tooltip to display if field is invalid.
   */
  tip: string;
  /*
   * Value to validate. Will be set as invalid if null, undefined or a empty string.
   */
  value?: any;
  /**
   * If set will be used instead of value to validate. Will be set to invalid if condition is false.
   */
  condition?: boolean;
}

export interface ModalHeadlessProps
  extends Omit<ModalOverlayProps, "onClose" | "onOpen"> {
  /**
   * State on whether modal is currently visible.
   */
  visible: boolean;
  /**
   * Set State on whether modal is currently visible.
   */
  setVisible: (visible: boolean) => any;

  onClose?: (...e: any) => any;

  onOpen?: (...e: any) => any;
}

/**
 * Modal component were the visibility state is managed externally in the parents state.
 */
export function ModalHeadless({
  visible,
  setVisible,
  ...overlayProps
}: ModalHeadlessProps) {
  let overlay = null;
  if (visible) {
    overlay = (
      <ModalOverlay
        {...overlayProps}
        onOpen={() => {
          if (overlayProps.onOpen) {
            overlayProps.onOpen();
          }
          setVisible?.(true);
        }}
        onClose={() => {
          if (overlayProps.onClose) {
            overlayProps.onClose();
          }
          setVisible?.(false);
        }}
      />
    );
  }

  return <>{overlay}</>;
}

// // extends Omit<ModalHeadlessProps, "visible" | "setVisible"> {
export interface ModalProps
  extends Omit<ModalHeadlessProps, "visible" | "setVisible"> {
  /**
   * Button React Node that will open Modal on click.
   */
  button?: React.ReactNode;
  /**
   * Button Props for `button` React Node that will open Modal on click.
   */
  buttonProps?: ButtonProps;
  /**
   * Disables Modal open button.
   */
  buttonDisabled?: boolean;
  /**
   * If no button is passed to Modal props, a button will automatically be created containing this label.
   */
  label?: string;
}

/**
 * Modal component were the visibility state is managed by a child button component.
 */
function Modal({
  button,
  buttonProps,
  buttonDisabled,
  label = "Click me",
  ...props
}: ModalProps) {
  const [visible, setVisible] = useState(false);
  if (!button) {
    return (
      <div>
        <ModalHeadless {...props} setVisible={setVisible} visible={visible} />
        <Button
          disabled={buttonDisabled}
          onClick={(ev) => {
            setVisible(true);
            ev.stopPropagation();
          }}
        >
          {label}
        </Button>
      </div>
    );
  }

  return (
    <>
      <ModalHeadless {...props} setVisible={setVisible} visible={visible} />
      <Button
        onClick={async (ev) => {
          if (props.onOpen) {
            await props.onOpen();
          }
          setVisible(true);
          ev.stopPropagation();
        }}
        variant="outlined"
        color="info"
        {...buttonProps}
      >
        {button}
      </Button>
    </>
  );
}
export default Modal;

export interface ModalOverlayProps {
  /**
   * Maximise Modal width and height.
   */
  maximised?: boolean;
  /**
   * Remove padding from Modal overlay.
   */
  noPadding?: boolean;
  /**
   * Icon that will be placed in the Modal's header
   */
  titleIcon?: React.ReactNode;
  /**
   * Title label that will be placed in the Modal's header
   */
  title: string;
  /**
   * Children of Modal overlay.
   */
  children?: React.ReactNode;
  /**
   * Automatically add close button to Modal header.
   */
  withCloseButton?: boolean;
  /**
   * TODO
   */
  activeOverlay?: boolean;
  /**
   * CSS Properties to be applied to modal element.
   */
  modalStyle?: React.CSSProperties;
  /**
   * Listen for escape keypress, and close modal when pressed.
   */
  escapeToCancel?: boolean;
  /**
   * Required fields to validate. To be used in conjunction with `actionButtons` validate `property`.
   */
  requiredFields?: RequiredField[];
  /**
   * Function that runs when Modal opens.
   */
  onOpen: (...e: any) => any;
  /**
   * Function that runs when Modal closes.
   */
  onClose: (...e: any) => any;
  /**
   * Buttons that will be added to Modal footer.
   */
  actionButtons?: ButtonDefinition[];
  /**
   * Key down event listners on Modal.
   */
  keyDownActions?: KeyDownDefinition[];
}

function ModalOverlay({
  maximised,
  noPadding,
  titleIcon,
  title = "Modal Title",
  children,
  withCloseButton,
  activeOverlay,
  modalStyle,
  actionButtons,
  keyDownActions,
  escapeToCancel,
  onClose,
  requiredFields,
}: ModalOverlayProps) {
  function validateFields(reqFields?: RequiredField[]) {
    const tipMessages: string[] = [];

    if (!reqFields) {
      return { tips: tipMessages, valid: tipMessages.length === 0 };
    }

    for (let i = 0; i < reqFields.length; i++) {
      const rField = reqFields[i];
      // @ts-expect-error
      if (rField.condition !== undefined) {
        // @ts-expect-error
        if (!rField.condition) {
          // @ts-expect-error
          tipMessages.push(rField.tip);
        }
        continue;
      }

      // Check if value is set
      // @ts-expect-error
      if (!rField.value === null || rField.value === undefined) {
        // @ts-expect-error
        tipMessages.push(rField.tip);
        continue;
      }

      // @ts-expect-error
      const rFieldType = typeof rField.value;
      // @ts-expect-error
      if (rFieldType === "string" && rField.value === "") {
        // @ts-expect-error
        tipMessages.push(rField.tip);
      }
    }

    return { tips: tipMessages, valid: tipMessages.length === 0 };
  }

  const [displayAlert, setDisplayAlert] = useState(false);
  const [alertMessage, setAlertMessage] = useState("");

  const validateResults = validateFields(requiredFields);

  useEffect(() => {
    function closeModal(e: KeyboardEvent) {
      if (e.keyCode === 27) {
        onClose(false);
      }
    }

    const removeListeners: { label: string; fn: (...e: any) => any }[] = [];

    if (escapeToCancel) {
      window.addEventListener("keydown", closeModal);
      removeListeners.push({ label: "keydown", fn: closeModal });
    }

    if (keyDownActions) {
      for (let i = 0; i < keyDownActions.length; i++) {
        const action = keyDownActions[i];

        const fn = async (e: KeyboardEvent) => {
          const eventTarget: any = e.target;

          // Check if event target matches keyboard action id
          // @ts-expect-error
          if (action.id !== undefined && action.id !== eventTarget.id) {
            return;
          }

          // @ts-expect-error
          if (e.code === action.code) {
            try {
              // @ts-expect-error
              const result = await action.fn();
              // @ts-expect-error
              if (!result?.error && action.closeModal) {
                onClose(false);
              }
              if (result?.error) {
                setAlertMessage(result?.msg);
                setDisplayAlert(true);
              }
            } catch (err) {
              if (err instanceof Error) {
                setAlertMessage(err.message);
                setDisplayAlert(true);
              } else {
                //TODO: HANDLE BAD ERROR
              }
            }
          }
        };

        window.addEventListener("keydown", fn);
        removeListeners.push({ label: "keydown", fn: fn });
      }
    }

    return () => {
      for (let i = 0; i < removeListeners.length; i++) {
        window.removeEventListener(
          // @ts-expect-error
          removeListeners[i].label,
          // @ts-expect-error
          removeListeners[i].fn
        );
      }
    };
  }, [escapeToCancel, onClose, keyDownActions]);

  let overlayClasses = "";
  let closeButton = null;
  if (withCloseButton) {
    closeButton = (
      <FlexBox
        className="modal-buttons"
        style={{ flexDirection: "column-reverse" }}
      >
        <div>
          <VscClose
            onClick={() => {
              onClose();
            }}
            className="auto-margin"
            style={{ marginRight: "8px" }}
          />
        </div>
      </FlexBox>
    );
  }

  if (activeOverlay) {
    overlayClasses += "clickable";
  }

  let buttons;
  if (actionButtons) {
    buttons = generateButtons(
      onClose,
      setDisplayAlert,
      setAlertMessage,
      actionButtons,
      validateResults
    );
  }

  let contentBodyStyle = {};
  if (!noPadding) {
    contentBodyStyle = {
      padding: "12px",
    };
  }

  return (
    <>
      <div className={"modal-overlay " + overlayClasses} />
      <div
        className={"modal-container " + overlayClasses}
        onClick={() => {
          if (activeOverlay) {
            onClose();
          }
        }}
      >
        <FlexBox tall>
          <div
            style={{
              display: "flex",
              width: "100%",
              justifyContent: "center",
              ...modalStyle,
            }}
            className="modal-body auto-margin"
            onClick={(e) => {
              e.stopPropagation();
            }}
          >
            <ContentPanel
              style={{
                maxHeight: "90vh",
                height: "100%",
                minWidth: "20vw",
                maxWidth: "80vw",
                overflowY: "auto",
                width: maximised ? "90vw" : "100%",
              }}
            >
              <ContentPanelTitle>
                <FlexBox style={{ maxWidth: "18px" }}>
                  <ContentPanelTitleIcon>
                    {titleIcon ? [titleIcon] : <VscDiffAdded />}
                  </ContentPanelTitleIcon>
                </FlexBox>
                <FlexBox>{title}</FlexBox>
                <FlexBox>{closeButton}</FlexBox>
              </ContentPanelTitle>
              <ContentPanelBody
                style={{ ...contentBodyStyle, overflow: "auto" }}
              >
                <FlexBox col gap>
                  {displayAlert ? (
                    <Alert
                      severity="error"
                      variant="filled"
                      onClose={() => {
                        setDisplayAlert(false);
                      }}
                    >
                      {alertMessage}
                    </Alert>
                  ) : null}
                  {children}
                </FlexBox>
              </ContentPanelBody>
              {buttons ? (
                <div>
                  <ContentPanelFooter>
                    <FlexBox
                      className="gap modal-buttons-container"
                      style={{ flexDirection: "row-reverse" }}
                    >
                      {buttons}
                    </FlexBox>
                  </ContentPanelFooter>
                </div>
              ) : null}
            </ContentPanel>
          </div>
        </FlexBox>
      </div>
    </>
  );
}

// export function ButtonDefinition(label, onClick, buttonProps, errFunc, closesModal, async, validate) {
//     return {
//         label: label,
//         onClick: onClick,
//         buttonProps: buttonProps,
//         errFunc: errFunc,
//         closesModal: closesModal,
//         async: async,
//         validate: validate
//     }
// }

export interface KeyDownDefinition {
  /**
   * Key code for add event listener that will trigger `fn`.
   */
  code: string;
  /**
   * Function that runs when action key is press.
   */
  fn: (...e: any) => any;
  /**
   * Callback function to run if fn throws an error.
   */
  errFunc: (...e: any) => any;
  /**
   * Closes modal on key down.
   */
  closeModal?: boolean;
  id?: string;
}
// KeyDownDefinition :
// code : Target Key Event
// fn : onClose function
// closeModal : Whether to close modal after fn()
// id : target element id to listen on. If undefined listener is global
// export function KeyDownDefinition(code, fn, errFunc, closeModal, targetElementID) {
//     return {
//         code: code,
//         fn: fn,
//         errFunc: errFunc,
//         closeModal: closeModal,
//         id: targetElementID,
//     }
// }

export interface ButtonDefinition {
  /**
   * Label to be placed in action button.
   */
  label: string;
  /**
   * Function that runs when action button is clicked.
   */
  onClick?: (...e: any) => any;
  /**
   * Button Props for this action button.
   */
  buttonProps?: ButtonProps;
  /**
   * Callback function to run if onClick throws an error.
   */
  errFunc?: (...e: any) => any;
  /**
   * Closes modal on click.
   */
  closesModal?: boolean;
  /**
   * If true, button will be disabled if any modal required fields are invalid.
   */
  validate?: boolean;
}

function generateButtons(
  onClose: (...e: any) => any,
  setDisplayAlert: React.Dispatch<React.SetStateAction<boolean>>,
  setAlertMessage: React.Dispatch<React.SetStateAction<string>>,
  actionButtons: ButtonDefinition[],
  validateResults: {
    tips: string[];
    valid: boolean;
  }
) {
  // label, onClick, classList, closesModal, async

  const out = [];
  for (let i = 0; i < actionButtons.length; i++) {
    const btn = actionButtons[i];

    const onClick = async () => {
      try {
        // @ts-expect-error
        if (btn.onClick) {
          // @ts-expect-error
          await btn.onClick();
        }

        // @ts-expect-error
        if (btn.closesModal) {
          onClose();
        } else {
          setAlertMessage("");
          setDisplayAlert(false);
        }
      } catch (e) {
        // @ts-expect-error
        if (btn.errFunc) {
          // @ts-expect-error
          await btn.errFunc();
        }

        if (e instanceof Error) {
          setAlertMessage(e.message);
        } else {
          //TODO: HANDLE BAD ERROR
        }
        setDisplayAlert(true);
      }
    };

    out.push(
      <Button
        variant="outlined"
        color="info"
        key={Array(5)
          .fill(0)
          .map(() =>
            "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789".charAt(
              Math.random() * 62
            )
          )
          .join("")}
        disabledTooltip={`${validateResults.tips.join(", ")}`}
        // @ts-expect-error
        disabled={!validateResults.valid && btn.validate}
        onClick={onClick}
        // @ts-expect-error
        {...btn.buttonProps}
      >
        {/* @ts-expect-error  */}
        <div>{btn.label}</div>
      </Button>
    );
  }

  return out;
}
