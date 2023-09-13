#!/bin/bash

# NAME: grub-menu.sh
# PATH: $HOME/bin
# DESC: Written for AU Q&A: https://askubuntu.com/q/1019213/307523
# DATE: Apr 5, 2018. Modified: May 7, 2018.

# $TERM variable may be missing when called via desktop shortcut
CurrentTERM=$(env | grep TERM)
if [[ $CurrentTERM == "" ]] ; then
    notify-send --urgency=critical "$0 cannot be run from GUI without TERM environment variable."
    exit 1
fi

AllMenusArr=()      # All menu options.
# Default for hide duplicate and triplicate options with (upstart) and (recovery mode)?
HideUpstartRecovery=false
if [[ $1 == short ]] ; then
    HideUpstartRecovery=true    # override default with first passed parameter "short"
elif [[ $1 == long ]] ; then
    HideUpstartRecovery=false   # override default with first passed parameter "long"
fi
SkippedMenuEntry=false  # Don't change this value, automatically maintained
InSubMenu=false     # Within a line beginning with `submenu`?
InMenuEntry=false   # Within a line beginning with `menuentry` and ending in `{`?
NextMenuEntryNo=0   # Next grub internal menu entry number to assign
# Major / Minor internal grub submenu numbers, ie `1>0`, `1>1`, `1>2`, etc.
ThisSubMenuMajorNo=0
NextSubMenuMinorNo=0
CurrTag=""          # Current grub internal menu number, zero based
CurrText=""         # Current grub menu option text, ie "Ubuntu", "Windows...", etc.
SubMenuList=""      # Only supports 10 submenus! Numbered 0 to 9. Future use.

while read -r line; do
    # Example: "           }"
    BlackLine="${line//[[:blank:]]/}" # Remove all whitespace
    if [[ $BlackLine == "}" ]] ; then
        # Add menu option in buffer
        if [[ $SkippedMenuEntry == true ]] ; then
            NextSubMenuMinorNo=$(( $NextSubMenuMinorNo + 1 ))
            SkippedMenuEntry=false
            continue
        fi
        if [[ $InMenuEntry == true ]] ; then
            InMenuEntry=false
            if [[ $InSubMenu == true ]] ; then
                NextSubMenuMinorNo=$(( $NextSubMenuMinorNo + 1 ))
            else
                NextMenuEntryNo=$(( $NextMenuEntryNo + 1 ))
            fi
        elif [[ $InSubMenu == true ]] ; then
            InSubMenu=false
            NextMenuEntryNo=$(( $NextMenuEntryNo + 1 ))
        else
            continue # Future error message?
        fi
        # Set maximum CurrText size to 68 characters.
        CurrText="${CurrText:0:67}"
        AllMenusArr+=($CurrTag "$CurrText")
    fi

    # Example: "menuentry 'Ubuntu' --class ubuntu --class gnu-linux --class gnu" ...
    #          "submenu 'Advanced options for Ubuntu' $menuentry_id_option" ...
    if [[ $line == submenu* ]] ; then
        # line starts with `submenu`
        InSubMenu=true
        ThisSubMenuMajorNo=$NextMenuEntryNo
        NextSubMenuMinorNo=0
        SubMenuList=$SubMenuList$ThisSubMenuMajorNo
        CurrTag=$NextMenuEntryNo
        CurrText="${line#*\'}"
        CurrText="${CurrText%%\'*}"
        AllMenusArr+=($CurrTag "$CurrText") # ie "1 Advanced options for Ubuntu"

    elif [[ $line == menuentry* ]] && [[ $line == *"{"* ]] ; then
        # line starts with `menuentry` and ends with `{`
        if [[ $HideUpstartRecovery == true ]] ; then
            if [[ $line == *"(upstart)"* ]] || [[ $line == *"(recovery mode)"* ]] ; then
                SkippedMenuEntry=true
                continue
            fi
        fi
        InMenuEntry=true
        if [[ $InSubMenu == true ]] ; then
            : # In a submenu, increment minor instead of major which is "sticky" now.
            CurrTag=$ThisSubMenuMajorNo">"$NextSubMenuMinorNo
        else
            CurrTag=$NextMenuEntryNo
        fi
        CurrText="${line#*\'}"
        CurrText="${CurrText%%\'*}"

    else
        continue    # Other stuff - Ignore it.
    fi

done < /boot/grub/grub.cfg

LongVersion=$(grub-install --version)
ShortVersion=$(echo "${LongVersion:20}")
DefaultItem=0

if [[ $HideUpstartRecovery == true ]] ; then
    MenuText="Menu No.     ----------- Menu Name -----------"
else
    MenuText="Menu No. --------------- Menu Name ---------------"
fi

while true ; do

    Choice=$(whiptail \
        --title "Use arrow, page, home & end keys. Tab toggle option" \
        --backtitle "Grub Version: $ShortVersion" \
        --ok-button "Display Grub Boot" \
        --cancel-button "Exit" \
        --default-item "$DefaultItem" \
        --menu "$MenuText" 24 76 16 \
        "${AllMenusArr[@]}" \
        2>&1 >/dev/tty)

    clear
    if [[ $Choice == "" ]]; then break ; fi
    DefaultItem=$Choice

    for (( i=0; i < ${#AllMenusArr[@]}; i=i+2 )) ; do
        if [[ "${AllMenusArr[i]}" == $Choice ]] ; then
            i=$i+1
            MenuEntry="menuentry '"${AllMenusArr[i]}"'"
            break
        fi
    done

    TheGameIsAfoot=false
    while read -r line ; do
        if [[ $line = *"$MenuEntry"* ]]; then TheGameIsAfoot=true ; fi
        if [[ $TheGameIsAfoot == true ]]; then
            echo $line
            if [[ $line = *"}"* ]]; then break ; fi
        fi
    done < /boot/grub/grub.cfg

    read -p "Press <Enter> to continue"

done

exit 0
