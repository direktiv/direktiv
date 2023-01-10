import MUIAlert, { AlertProps as MUIAlertProps } from '@mui/material/Alert';
import { styled } from '@mui/material/styles';
import { VscWarning } from 'react-icons/vsc';
import './style.css';


export interface AlertProps extends MUIAlertProps{
    /**
     * Sets width to 100%.
     */
    grow?: boolean;
    /**
    * The severity of the alert. This defines the color and icon used.
    */
    severity: 'success' | 'info' | 'warning' | 'error'
}

const DirektivAlert = styled(MUIAlert, {
    shouldForwardProp: (prop) => prop !== 'grow',
})<AlertProps>(({ theme, variant, severity, grow }) => ({
    ...(grow && {
        width: "100%",
    }),
    ...(severity === "info" && variant === undefined && {
        backgroundColor: "#cfd5de",
        color: "#566875",
        ".MuiAlert-icon": {
            color: "#566875",
        }
    }),
    ...(severity === "error" && variant === undefined && {
    }),
}));

/**
 * Component that displays a short simple message in a Alert container. 
 */
function Alert({
    severity = "info",
    ...props
  }: AlertProps) {
    return (
        <DirektivAlert iconMapping={{
            error: <VscWarning fontSize="inherit" />
          }}
          {...props} severity={severity}/>
    )
}

export default Alert