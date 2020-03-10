import Vue from 'vue';
import Vuex from './vuex'


const store = new Vuex.Store({
  state: {
    name: "suke",
    obj: {
      name: "test"
    },
    isDev: false
  }
})


export default store
