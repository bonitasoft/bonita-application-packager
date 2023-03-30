ARG BONITA_BASE_IMAGE=bonita:latest
FROM $BONITA_BASE_IMAGE

ARG CUSTOM_APPLICATION_RESOURCES_FOLDER="resources"
ARG BONITA_CUSTOM_APPLICATION_FOLDER="my-application"
ARG BONITA_WEB_INF_CLASSES_PATH="/opt/bonita/server/webapps/bonita/WEB-INF/classes"

# Copy custom bonita application resources (e.g. .zip and .bconf) to the dedicated bundle folder
COPY --chown=bonita:bonita ${CUSTOM_APPLICATION_RESOURCES_FOLDER}/* ${BONITA_WEB_INF_CLASSES_PATH}/${BONITA_CUSTOM_APPLICATION_FOLDER}/
