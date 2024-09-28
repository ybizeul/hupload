import { ActionIcon, Affix, Tooltip } from "@mantine/core";
import { IconArrowLeft, IconLogout } from "@tabler/icons-react";
import { Link, useMatch } from "react-router-dom";
import { H } from "@/APIClient";
import { useAuthContext } from "@/AuthContext";
import { useTranslation } from "react-i18next";

export function Haffix() {
    const { t } = useTranslation();

    const location = useMatch('/:share')
    const { setAuthInfo } = useAuthContext()

    return(
        <>
            <Affix position={{ bottom: 50 }} w="100%" ta="center">
                <Tooltip withArrow arrowOffset={10} arrowSize={4} label={t("shares")}>
                    <ActionIcon size="lg" mr="lg" disabled={location?.params.share == 'shares'} component={Link} to="/shares" color="blue" radius="xl"><IconArrowLeft style={{ width: '70%', height: '70%' }} /></ActionIcon>
                </Tooltip>
                <Tooltip withArrow arrowOffset={10} arrowSize={4} label={t("logout")}>
                    <ActionIcon size="lg" variant="light" color="red" radius="xl" onClick={() => { setAuthInfo(null);H.logoutNow(); window.location.href='/'}}><IconLogout style={{ width: '70%', height: '70%' }} /></ActionIcon>
                </Tooltip>
            </Affix>
        </>
    )
}