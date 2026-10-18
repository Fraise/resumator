import { mount } from 'svelte'

import './style.css'
import App from './App.svelte'

const target = document.getElementById('app')
const app = target ? mount(App, { target }) : null

export default app
