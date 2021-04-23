<template>
    <v-app>
    <v-app-bar
      app
      color="black"
      dark
    >
      <div class="d-flex align-center">
        <v-toolbar-title>
            Paul The Alien
        </v-toolbar-title>
      </div>

      <v-spacer></v-spacer>

      <v-btn
        href=""
        target="_blank"
        text
      >
          <span class="mr-2">
              <v-img
                  :src="require('./assets/giphy-watermark.png')"
                  class="mr-2"
                  contain
              />
          </span>
      </v-btn>
      <v-btn
        href="https://github.com/Spazzy757/paul"
        target="_blank"
        text
      >
        <span class="mr-2">Latest Release</span>
        <v-icon>mdi-github</v-icon>
      </v-btn>
    </v-app-bar>

    <v-main>
      <Paul/>
    </v-main>
  </v-app>
</template>

<script>
 import Paul from './components/Paul';

 export default {
     name: 'App',

     components: {
         Paul,
     },

     data: () => ({
         //
     }),
 };
</script>

<style lang="scss" >
</style>
