import React from 'react';
import SecretsPanel from './secrets-panel';
import RegistriesPanel from './registries-panel';
import FlexBox from '../../components/flexbox';
import BroadcastConfigurationsPanel from './broadcast-panel';
import VariablesPanel from './variables-panel';

function ExamplePage(props) {
    return(
        <FlexBox id="settings-page" className="col gap" style={{ paddingRight: "8px" }}>
            <FlexBox className="gap">
                <FlexBox className="col gap" style={{ maxWidth: "380px" }}>
                    <FlexBox style={{ minWidth: "380px", maxWidth: "380px" }}>
                        <SecretsPanel />
                    </FlexBox>
                    <div style={{ minWidth: "380px", maxWidth: "380px" }}>
                        <RegistriesPanel />
                    </div>
                </FlexBox>
                <FlexBox>
                    <BroadcastConfigurationsPanel />
                </FlexBox>
            </FlexBox>
            <FlexBox>
                <VariablesPanel />
            </FlexBox>
        </FlexBox>
    )
}

export default ExamplePage;