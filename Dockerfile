FROM ubuntu
COPY ./bin/Gimulator /app/gimulator
WORKDIR /app
CMD ["./gimulator", "-ip=localhost:3030", "-config-file=/configs/roles.yaml"]
