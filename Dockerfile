ARG BONITA_IMAGE_NAME=bonita
ARG BONITA_IMAGE_VERSION=latest
FROM $BONITA_IMAGE_NAME:$BONITA_IMAGE_VERSION

ARG CUSTOM_APPLICATION_RESOURCES_FOLDER="resources"
ARG BONITA_CUSTOM_APPLICATION_FOLDER="my-application"

ARG BONITA_WEB_INF_CLASSES_PATH="/opt/bonita/server/webapps/bonita/WEB-INF/classes"
ARG BONITA_PAGES_PATH="${BONITA_WEB_INF_CLASSES_PATH}/org/bonitasoft/web/page"
ARG BONITA_APPLICATIONS_PATH="${BONITA_WEB_INF_CLASSES_PATH}/org/bonitasoft/web/application"

# Copy custom bonita application resources (e.g. .zip and .bconf) to the dedicated bundle folder
COPY --chown=bonita:bonita ${CUSTOM_APPLICATION_RESOURCES_FOLDER}/* ${BONITA_WEB_INF_CLASSES_PATH}/${BONITA_CUSTOM_APPLICATION_FOLDER}/

# Remove unecessary provided Bonita generic pages and applications
RUN rm ${BONITA_PAGES_PATH}/*.zip ${BONITA_APPLICATIONS_PATH}/*.zip
