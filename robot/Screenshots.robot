*** Settings ***
Library   Browser

*** Variables ***
${theme}  light

*** Test Cases ***
Screenshot Login Page
    FOR    ${theme}    IN    light  dark
        New Context      colorScheme=${theme}  viewport={'width': 800, 'height': 589}
        New Page         http://localhost:5173/
        Take Screenshot  ${CURDIR}/../readme_images/login-${theme}.png  crop={'x': 0, 'y': 0, 'width': 800, 'height': 589}
    END

Screenshot Home Page
    FOR    ${theme}    IN    light  dark
        New Context          colorScheme=${theme}  viewport={'width': 800, 'height': 589}
        New Page             http://localhost:5173/
        Fill Text            id=username  admin
        Fill Text            id=password  hupload
        Click                "Login"
        Wait For Load State  networkidle
        Take Screenshot      ${CURDIR}/../readme_images/shares-${theme}.png  crop={'x': 0, 'y': 0, 'width': 800, 'height': 589}

        Click            css=\#kuva-yibi-bata \#edit
        Sleep            0.5 second
        Click            css=\#showEditor
        Sleep            0.5 second
        Take Screenshot  ${CURDIR}/../readme_images/properties-${theme}.png  crop={'x': 0, 'y': 216, 'width': 800, 'height': 345}
        Click            css=\#preview
        Take Screenshot  ${CURDIR}/../readme_images/properties-preview-${theme}.png  crop={'x': 0, 'y': 216, 'width': 800, 'height': 345}


        Click            css=\#kuva-yibi-bata \#edit

        Click                "xube-suwe-hybe"
        Wait For Load State  networkidle
        Take Screenshot      ${CURDIR}/../readme_images/share-${theme}.png  crop={'x': 0, 'y': 0, 'width': 800, 'height': 589}


    END
