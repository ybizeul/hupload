import { ActionIcon, Affix, Tooltip } from "@mantine/core";
import { IconArrowLeft, IconLogout } from "@tabler/icons-react";
import { Link, useMatch, useNavigate } from "react-router-dom";
import { H } from "@/APIClient";
import { useAuthContext } from "@/AuthContext";
import { useTranslation } from "react-i18next";

export function Haffix() {
    const { t } = useTranslation();
    const navigate = useNavigate();
    const location = useMatch('/:share')
    const { authInfo, check } = useAuthContext()

    const logout = () => {
        if (authInfo?.logoutUrl) {
            document.cookie = "hupload" +'=; Path=/; Expires=Thu, 01 Jan 1970 00:00:01 GMT;';
            window.location.href = authInfo.logoutUrl
            return
        }
        H.logoutNow()
        check && check()
        navigate('/')
    }
    return(
        <>
            <Affix position={{ bottom: 50 }} w="100%" ta="center">
                <Tooltip withArrow arrowOffset={10} arrowSize={4} label={t("shares")}>
                    <ActionIcon size="lg" mr="lg" disabled={location?.params.share == 'shares'} component={Link} to="/shares" color="blue" radius="xl"><IconArrowLeft style={{ width: '70%', height: '70%' }} /></ActionIcon>
                </Tooltip>
                <Tooltip withArrow arrowOffset={10} arrowSize={4} label={t("logout")}>
                    <ActionIcon size="lg" variant="light" color="red" radius="xl" onClick={() => { logout() }}><IconLogout style={{ width: '70%', height: '70%' }} /></ActionIcon>
                </Tooltip>
            </Affix>
        </>
    )
}