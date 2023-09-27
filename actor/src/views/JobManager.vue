<script setup lang="ts">
import {NButton, NDataTable, NSpace, NTag, useMessage} from "naive-ui"
import {h, onMounted, ref} from "vue";
import {getJobList, runJob, runTask, stopJob} from "@/request/remote"

const message = useMessage()
const columns = [
  {title: 'jobId', key: 'name', ellipsis: true},
  {
    title: 'type', key: 'type', ellipsis: true, render(row: any) {
      return row.type === 1 ? "常驻" : "定时"
    }
  },
  {
    title: 'run', key: 'openRun', ellipsis: true, render(row: any) {
      return h(NTag, {
        bordered: false,
        type: row.openRun ? "info" : "default",
      }, {default: () => row.openRun ? "开启" : "关闭"})
    }
  },
  {
    title: 'status', key: 'status', ellipsis: true, render(row: any) {
      return h(NTag, {
            bordered: false,
            type: row.status === 1 ? "success" : "warning"
          },
          {default: () => row.status === 1 ? "运行" : "暂停"}
      )
    }
  },
  {
    title: 'opt', key: 'opt', ellipsis: true, render(row: any) {
      let jobButton = [h(
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
      )]
      let taskButton = [
        h(NButton,
            {
              strong: true,
              tertiary: true,
              type: "primary",
              size: "small",
              onClick: () => runTask(row.name).then(r => {
                message.success(r.data.message)
                return getData()
              })
            },
            {default: () => "运行一次"}
        )
      ]
      if (row.type === 1) {
        return h(NSpace,
            {},
            {default: () => jobButton})
      } else {

        return h(NSpace,
            {},
            {default: () => taskButton})
      }
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