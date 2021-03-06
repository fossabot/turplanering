/** @license
 *
 *  Turplanering
 *  Copyright (C) 2020 Emil Gedda
 *
 *  This program is free software: you can redistribute it and/or modify
 *  it under the terms of the GNU Affero General Public License as published by
 *  the Free Software Foundation, either version 3 of the License, or
 *  (at your option) any later version.
 *
 *  This program is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *  GNU Affero General Public License for more details.
 *
 *  You should have received a copy of the GNU Affero General Public License
 *  along with this program. If not, see <https://www.gnu.org/licenses/>.
 */
import React from 'react'
import { render } from 'react-dom'
import { App } from './components/App'

const getEnvironment = () => {
    const browser = {
        hasTouch: 'ontouchstart' in window || navigator.maxTouchPoints > 0,
    }
    if (process.env.NODE_ENV == 'production') {
        return {
            apiURL: 'localhost:8080',
            environment: 'production',
            browser,
        }
    }
    return {
        apiURL: 'localhost:8080',
        environment: 'development',
        browser,
    }
}

render(
    <React.StrictMode>
        <App env={getEnvironment()} />
    </React.StrictMode>,
    document.getElementById('root') as HTMLElement
)
