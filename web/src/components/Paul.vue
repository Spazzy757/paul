<template>
  <v-container>
    <v-row class="text-center">
      <v-col cols="12">
        <v-img
          :src="require('../assets/paul.png')"
          class="my-3"
          contain
          height="200"
        />
      </v-col>

      <v-col class="mb-4">
        <h1 class="display-2 font-weight-bold mb-3">
          Say Hello to Paul The Alien
        </h1>

        <p class="subheading font-weight-regular">
          The awesome GitHub Bot here to make your workflow fun
        </p>
      </v-col>

      <v-col
        class="mb-5"
        cols="12"
      >
        <h2 class="headline font-weight-bold mb-3">
          Install Paul on your Repo via GitHub Apps
        </h2>

        <v-row justify="center">
            <a
                href="https://github.com/apps/paul-the-alien"
                class="subheading mx-3"
                target="_blank"
            >
            Installation Instructions
          </a>
        </v-row>
      </v-col>

      <v-col
        class="mb-5"
        cols="12"
      >
        <h2 class="headline font-weight-bold mb-3">
            Have an Idea or Having Problems? Log an Issue!
        </h2>

        <v-row justify="center">
            <a
                href="https://github.com/Spazzy757/paul/issues"
                class="subheading mx-3"
                target="_blank"
            >
            Pauls Issue Page
          </a>
        </v-row>
      </v-col>

      <v-col
        class="mb-5"
        cols="12"
      >
        <h2 class="headline font-weight-bold mb-3">
          Paul is a Free and a Open Source service which is currently in Alpha
        </h2>

      </v-col>
    </v-row>
  </v-container>
</template>

<script>
  export default {
    name: 'Paul',
  }
</script>
