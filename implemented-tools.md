

1 - CREATE USER
curl http://localhost:8080/mcp/v1/run -H "content-type: application/json" -d '{"tool":"create_user","input":{"environment_id":"5402f462-7316-4699-80db-c063d55b9aac","username":"aadoe","email":"jdoe@example.com"}}'

2 - GET USER
curl http://localhost:8080/mcp/v1/run -H "content-type: application/json" -d '{"tool":"get_user","input":{"environment_id":"5402f462-7316-4699-80db-c063d55b9aac","id":"6ccd553b-aa7b-4674-9f4c-be402cbe1975"}}'

3 - UPDATE USER
curl http://localhost:8080/mcp/v1/run -H "content-type: application/json" -d '{"tool":"update_user","input":{"environment_id":"5402f462-7316-4699-80db-c063d55b9aac","id":"bb2d8afe-222c-4433-be0b-71ef1a4f5ee7","username":"new_aadoe","email":"new_aadoe@example.com"}}'

4 - DELETE USER
curl http://localhost:8080/mcp/v1/run -H "content-type: application/json" -d '{"tool":"delete_user","input":{"environment_id":"5402f462-7316-4699-80db-c063d55b9aac","id":"6ccd553b-aa7b-4674-9f4c-be402cbe1975"}}'

5 - PASSWORD STATE
curl http://localhost:8080/mcp/v1/run -H "content-type: application/json" -d '{"tool":"get_user_password_state","input":{"environment_id":"5402f462-7316-4699-80db-c063d55b9aac","id":"6ccd553b-aa7b-4674-9f4c-be402cbe1975"}}'

6 - RESET PASSWORD
curl http://localhost:8080/mcp/v1/run -H "content-type: application/json" -d '{"tool":"reset_user_password","input":{"environment_id":"5402f462-7316-4699-80db-c063d55b9aac","id":"6ccd553b-aa7b-4674-9f4c-be402cbe1975","password":"new_password"}}'

7 - UNLOCK PASSWORD
curl http://localhost:8080/mcp/v1/run -H "content-type: application/json" -d '{"tool":"unlock_user_password","input":{"environment_id":"5402f462-7316-4699-80db-c063d55b9aac","id":"6ccd553b-aa7b-4674-9f4c-be402cbe1975"}}'

8 - ADD USER TO  GROUP
curl http://localhost:8080/mcp/v1/run -H "content-type: application/json" -d '{"tool":"add_user_to_group","input":{"environment_id":"5402f462-7316-4699-80db-c063d55b9aac","user_id":"bb2d8afe-222c-4433-be0b-71ef1a4f5ee7","group_id":"4537e89d-e1cb-48fd-8ef7-2e9a9e0ebf80"}}'

9 - REMOVE USER FROM GROUP
curl http://localhost:8080/mcp/v1/run -H "content-type: application/json" -d '{"tool":"remove_user_from_group","input":{"environment_id":"5402f462-7316-4699-80db-c063d55b9aac","user_id":"bb2d8afe-222c-4433-be0b-71ef1a4f5ee7","group_id":"4537e89d-e1cb-48fd-8ef7-2e9a9e0ebf80"}}'

10 - CREATE POPULATION
curl http://localhost:8080/mcp/v1/run -H "content-type: application/json" -d '{"tool":"create_population","input":{"environment_id":"5402f462-7316-4699-80db-c063d55b9aac","name":"Test Population","description":"A test population for user management"}}'

11 - GET POPULATION
curl http://localhost:8080/mcp/v1/run -H "content-type: application/json" -d '{"tool":"get_population","input":{"environment_id":"5402f462-7316-4699-80db-c063d55b9aac","id":"fd48b84e-aade-4088-993e-6f6eb5560e37"}}'

12 - DELETE POPULATION
curl http://localhost:8080/mcp/v1/run -H "content-type: application/json" -d '{"tool":"delete_population","input":{"environment_id":"5402f462-7316-4699-80db-c063d55b9aac","id":"fd48b84e-aade-4088-993e-6f6eb5560e37"}}'

13 - GET POPULATIONS
curl http://localhost:8080/mcp/v1/run -H "Content-Type: application/json" -d '{"tool":"get_environment_populations","input":{"environment_id":"5402f462-7316-4699-80db-c063d55b9aac"}}'

14 - CREATE GROUP
curl http://localhost:8080/mcp/v1/run -H "content-type: application/json" -d '{"tool":"create_group","input":{"environment_id":"5402f462-7316-4699-80db-c063d55b9aac","name":"MCP Test Group","description":"A test group created via MCP server"}}'

15 - DELETE GROUP
curl http://localhost:8080/mcp/v1/run -H "content-type: application/json" -d '{"tool":"delete_group","input":{"environment_id":"5402f462-7316-4699-80db-c063d55b9aac","id":"44db33da-8a5d-43d9-b9ca-b89163769565"}}'

16 - GET GROUP
curl http://localhost:8080/mcp/v1/run -H "content-type: application/json" -d '{"tool":"get_group","input":{"environment_id":"5402f462-7316-4699-80db-c063d55b9aac","id":"bdcbca7c-781f-4329-9c4b-5a989fac32a4"}}'

17 - CREATE ENVIRONMENT
curl http://localhost:8080/mcp/v1/run -H "Content-Type: application/json" -d '{"tool":"create_environment","input":{"name":"MCP Test Env","description":"A test environment created via MCP server"}}'

18 - DELETE ENVIRONMENT
curl http://localhost:8080/mcp/v1/run -H "Content-Type: application/json" -d '{"tool":"delete_environment","input":{"id":"538ea7db-5813-437a-b3ad-cfb9b8b4bd11"}}'

19 - UPDATE ENVIRONMENT STATUS
curl http://localhost:8080/mcp/v1/run -H "Content-Type: application/json" -d '{"tool":"update_environment_status","input":{"id":"538ea7db-5813-437a-b3ad-cfb9b8b4bd11","status":"DELETE_PENDING"}}'

20 - GET LICENSES
curl http://localhost:8080/mcp/v1/run -H "Content-Type: application/json" -d '{"tool":"get_licenses","input":{}}'

21 - GET ENVIRONMENTS
curl http://localhost:8080/mcp/v1/run -H "Content-Type: application/json" -d '{"tool":"get_environments","input":{}}'

22 - GET ENVIRONMENT
curl http://localhost:8080/mcp/v1/run -H "Content-Type: application/json" -d '{"tool":"get_environment","input":{"id":"d907b33d-f383-45de-a132-a9108149b184"}}'

23 - GET ENVIRONMENT BOM
curl http://localhost:8080/mcp/v1/run -H "Content-Type: application/json" -d '{"tool":"get_environment_bom","input":{"id":"d907b33d-f383-45de-a132-a9108149b184"}}'

24 - GET LICENSE
curl http://localhost:8080/mcp/v1/run -H "Content-Type: application/json" -d '{"tool":"get_license","input":{"organization_id":"0a2ff87c-9479-42f9-ae97-6f44d1974be0","license_id":"10366220-58cc-49bd-8ee2-fdacbea62a85"}}'

25 - UPDATE GROUP
    curl -sS -X POST http://localhost:8080/mcp/v1/run \
      -H "Content-Type: application/json" \
      -d '{
        "tool": "update_group",
        "input": {
          "environment_id": "d907b33d-f383-45de-a132-a9108149b184",
          "id": "bdcbca7c-781f-4329-9c4b-5a989fac32a4",
          "name": "Upated MCP Group Name",
          "description": "Updated description"
        }
      }'

GET ORGANIZATION CAPABILITIES
GET ENVIRONMENT CAPABILITIES
USER PASSWORD FORCE CHANGE
USER PASSWORD RESEND RECOVERY CODE
GET USER POPULATIONS
UPDATE USER POPULATION
