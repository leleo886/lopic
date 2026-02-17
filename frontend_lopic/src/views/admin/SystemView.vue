<template>
  <div class="system-view">
    <div class="system-content card">
      <a-spin :spinning="loading && !isDataLoaded" :tip="t('system.loading')">
        <a-form
          v-if="isDataLoaded"
          :model="systemSettings"
          :rules="systemRules"
          ref="systemFormRef"
          @finish="handleSaveSettings"
        >
          
          <a-collapse v-model:activeKey="activeTabKey">
            <a-collapse-panel key="general" :header="t('system.general')">
              <a-form-item name="General.MaxThumbSize" :label="t('system.maxThumbSize')">
                <a-input-number v-model:value="systemSettings.General.MaxThumbSize" :min="300" :max="2048" />
                <span class="form-item-help">{{ t('system.pixels') }}</span>
              </a-form-item>
              <a-form-item name="General.RegisterEnabled" :label="t('system.registerEnabled')">
                <a-switch v-model:checked="systemSettings.General.RegisterEnabled" />
              </a-form-item>
              <a-form-item name="General.MaxTags" :label="t('system.maxTags')">
                <a-input-number v-model:value="systemSettings.General.MaxTags" :min="0" />
                <span class="form-item-help">{{ t('system.maxTagsHelp') }}</span>
              </a-form-item>
            </a-collapse-panel>

            <a-collapse-panel key="mail" :header="t('system.mail')">
              <a-form-item name="Mail.Enabled" :label="t('system.mailService')">
                <a-switch v-model:checked="systemSettings.Mail.Enabled" />
                <span class="form-item-help">
                  {{ t('system.mailTooltip') }}
                </span>
              </a-form-item>
              <a-form-item name="Mail.ServerAddress" :label="t('system.serverAddress')">
                <a-input v-model:value="systemSettings.Mail.ServerAddress" :placeholder="t('system.enterServerAddress')" />
              </a-form-item>
              <a-form-item name="Mail.SMTP.From" :label="t('system.senderEmail')">
                <a-input v-model:value="systemSettings.Mail.SMTP.From" :placeholder="t('system.enterSenderEmail')" />
              </a-form-item>
              <a-form-item name="Mail.SMTP.FromName" :label="t('system.senderName')">
                <a-input v-model:value="systemSettings.Mail.SMTP.FromName" :placeholder="t('system.enterSenderName')" />
              </a-form-item>
              <a-form-item name="Mail.SMTP.Host" :label="t('system.smtpServer')">
                <a-input v-model:value="systemSettings.Mail.SMTP.Host" :placeholder="t('system.enterSmtpServer')" />
              </a-form-item>
              <a-form-item name="Mail.SMTP.Port" :label="t('system.smtpPort')">
                <a-input-number v-model:value="systemSettings.Mail.SMTP.Port" :min="1" :max="65535" />
              </a-form-item>
              <a-form-item name="Mail.SMTP.Username" :label="t('system.smtpUsername')">
                <a-input v-model:value="systemSettings.Mail.SMTP.Username" :placeholder="t('system.enterSmtpUsername')" />
              </a-form-item>
              <a-form-item name="Mail.SMTP.Password" :label="t('system.smtpPassword')">
                <a-input-password v-model:value="systemSettings.Mail.SMTP.Password" :placeholder="t('system.enterSmtpPassword')" />
              </a-form-item>
            </a-collapse-panel>

            <a-collapse-panel key="gallery" :header="t('system.gallery')">
              <a-form-item name="Gallery.Title" :label="t('system.galleryTitle')">
                <a-input v-model:value="systemSettings.Gallery.Title" :placeholder="t('system.enterGalleryTitle')" />
                <span class="form-item-help">
                  {{ t('system.galleryTitleTooltip') }}
                </span>
              </a-form-item>
              <a-form-item name="Gallery.BackgroundImage" :label="t('system.galleryBackgroundImage')">
                <a-input v-model:value="systemSettings.Gallery.BackgroundImage" :placeholder="t('system.enterGalleryBackgroundImage')" />
                <span class="form-item-help">
                  {{ t('system.galleryBackgroundImageTooltip') }}
                </span>
              </a-form-item>
              <a-form-item name="Gallery.CustomContent" :label="t('system.galleryCustomContent')">
                <a-textarea
                  v-model:value="systemSettings.Gallery.CustomContent"
                  :placeholder="t('system.enterGalleryCustomContent')"
                  :rows="6"
                />
                <span class="form-item-help">
                  {{ t('system.galleryCustomContentTooltip') }}
                </span>
              </a-form-item>
            </a-collapse-panel>
          
          </a-collapse>
          
          <a-form-item>
            <a-button type="primary" html-type="submit" :loading="loading" class="save-button">
              {{ t('system.saveSettings') }}
            </a-button>
          </a-form-item>
        </a-form>
        <div v-else class="loading-placeholder">
          <p>{{ t('system.loadingSettings') }}</p>
        </div>
      </a-spin>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue';
import { message } from 'ant-design-vue';
import { systemApi } from '../../api/services';
import type { SystemSettings } from '../../types/api';
import { getErrorMessage } from '../../types/errorMessages';
import { useI18n } from 'vue-i18n';

const { t } = useI18n();

// 加载状态
const loading = ref(false);

// 数据加载状态
const isDataLoaded = ref(false);

// 表单引用
const systemFormRef = ref();

// 激活的标签
const activeTabKey = ref(['general']);

// 系统设置
const systemSettings = reactive<SystemSettings>({
  General: {
    MaxThumbSize: 800,
    RegisterEnabled: false,
    MaxTags: 0,
  },
  Mail: {
    Enabled: false,
    ServerAddress: '',
    SMTP: {
      Host: '',
      Port: 587,
      Username: '',
      Password: '',
      From: '',
      FromName: ''
    }
  },
  Gallery: {
    Title: 'Gallery',
    BackgroundImage: '',
    CustomContent: ''
  }
});

// 表单验证规则
const systemRules = {
  'Mail.SMTP.From': [
    { required: systemSettings.Mail.Enabled, message: () => t('system.enterSenderEmail'), trigger: 'blur' },
  ],
  'Mail.SMTP.FromName': [
    { required: systemSettings.Mail.Enabled, message: () => t('system.enterSenderName'), trigger: 'blur' },
  ],
  'Mail.SMTP.Host': [
    { required: systemSettings.Mail.Enabled, message: () => t('system.enterSmtpServer'), trigger: 'blur' },
  ],
  'Mail.SMTP.Port': [
    { required: systemSettings.Mail.Enabled, message: () => t('system.enterSmtpPort'), trigger: 'blur' },
  ],
  'Mail.SMTP.Username': [
    { required: systemSettings.Mail.Enabled, message: () => t('system.enterSmtpUsername'), trigger: 'blur' },
  ],
  'Mail.SMTP.Password': [
    { required: systemSettings.Mail.Enabled, message: () => t('system.enterSmtpPassword'), trigger: 'blur' },
  ],
};

// 获取系统设置
const fetchSystemSettings = async () => {
  try {
    loading.value = true;
    const response = await systemApi.getSystemInfo();
    
    // 直接使用后端返回的数据，因为前端类型定义已经与后端一致
    Object.assign(systemSettings, response.data);
    isDataLoaded.value = true;
  } catch (error: any) {
    const errorCode = error.response?.data?.code;
    const errorMessage = errorCode ? getErrorMessage(errorCode) : t('system.fetchFailed');
    message.error(errorMessage);
    isDataLoaded.value = true;
  } finally {
    loading.value = false;
  }
};

// 保存系统设置
const handleSaveSettings = async () => {
  if (!systemFormRef.value) return;
  
  try {
    await systemFormRef.value.validate();
    loading.value = true;
    
    // 直接使用前端的 systemSettings 对象，因为前端类型定义已经与后端一致
    await systemApi.updateSystemInfo(systemSettings);
    message.success(t('system.saveSuccess'));
  } catch (error: any) {
    const errorCode = error.response?.data?.code;
    const errorMessage = errorCode ? getErrorMessage(errorCode) : t('system.saveFailed');
    message.error(errorMessage);
  } finally {
    loading.value = false;
  }
};

// 初始化
onMounted(() => {
  fetchSystemSettings();
});
</script>

<style scoped>
.system-view {
  width: 100%;
}

.system-content {
  padding: var(--spacing-lg);
}

.save-button {
  margin-top: var(--spacing-lg);
  background-color: var(--primary-color);
  border-color: var(--primary-color);
}

.save-button:hover {
  background-color: var(--primary-dark);
  border-color: var(--primary-dark);
}

.form-item-help {
  margin-left: var(--spacing-sm);
  color: var(--text-secondary);
  font-size: var(--font-size-sm);
}

.loading-placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--spacing-xl);
  color: var(--text-secondary);
  min-height: 300px;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .system-view {
    width: 100%;
  }
  
  .system-content {
    padding: var(--spacing-md);
  }
  
  .form-item-help {
    display: block;
    margin-left: 0;
    margin-top: var(--spacing-xs);
    font-size: var(--font-size-xs);
  }
  
  :deep(.ant-form-item) {
    margin-bottom: var(--spacing-md);
  }
  
  :deep(.ant-form-item-label) {
    padding-bottom: var(--spacing-xs);
  }
  
  :deep(.ant-collapse-header) {
    font-size: var(--font-size-base);
    padding: var(--spacing-sm) var(--spacing-md);
  }
  
  :deep(.ant-collapse-content-box) {
    padding: var(--spacing-md);
  }
  
  .save-button {
    width: 100%;
  }
  
  .loading-placeholder {
    min-height: 200px;
    padding: var(--spacing-lg);
  }
}
</style>
