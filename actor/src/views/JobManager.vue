<script setup lang="ts">
import {NButton, NDataTable, NSpace} from "naive-ui"
import {h, onMounted, ref} from "vue";
import {getJobList, runJob, stopJob} from "@/request/remote"

const columns = [
  {title: 'jobId', key: 'name', ellipsis: true},
  {
    title: 'run', key: 'openRun', ellipsis: true, render(row: any) {
      return row.openRun === 1 ? "关闭" : "开启"
    }
  },
  {
    title: 'status', key: 'status', ellipsis: true, render(row: any) {
      return row.status === 1 ? "运行" : "暂停"
    }
  },
  {
    title: 'opt', key: 'opt', ellipsis: true, render(row: any) {
      return [h(
          NButton,
          {
            strong: true,
            tertiary: true,
            type: "error",
            size: "small",
            onClick: () => stopJob(row.name).then(r => getData())
          },
          {default: () => "停止"}
      ), h(
          NButton,
          {
            strong: true,
            tertiary: true,
            type: "primary",
            size: "small",
            onClick: () => runJob(row.name).then(r => getData())
          },
          {default: () => "run"}
      )
      ]
    }
  }
]
const data = ref([])

onMounted(() => {
  getData()
})

async function getData() {
  let resp = await getJobList()
  data.value = resp.data.message
  console.log(resp.data.message)
}
</script>
<template>

  <n-space>
    <n-button @click="getData">F5</n-button>
    <n-data-table
        :columns="columns"
        :data="data"

        :bordered="true"
    />
  </n-space>
  <!--        :pagination="pagination"-->
</template>