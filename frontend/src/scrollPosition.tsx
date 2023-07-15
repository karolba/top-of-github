import { h } from "dom-chef"

interface ScrollState {
    position: {
        x: number
        y: number
    }
}

// Save scroll position before navigating to a new page.
// Has to be called before each page navigation.
export function saveScrollPosition() {
    const scrollPosition: ScrollState = {
        position: {
            y: window.pageYOffset || document.documentElement.scrollTop || document.body.scrollTop,
            x: window.pageXOffset || document.documentElement.scrollLeft || document.body.scrollLeft
        }
    }
    history.replaceState(scrollPosition, '');
}

// Restore scroll position when navigating to us from somewhere else in the history stack
export function restoreScrollPosition() {
    if(window.history?.state && 'position' in window.history?.state) {
        const state: ScrollState = window.history.state
        setTimeout(() => {
            window.scrollTo({
                behavior: "instant",
                left: state.position.x,
                top: state.position.y
            })
        }, 0);
    }
}

// An element that's identical to the <a> element, but saves the scroll position before navigation
export function ScrollPositionSavingLink(props: any): JSX.Element {
    return <a onClick={saveScrollPosition} {...props}></a>
}