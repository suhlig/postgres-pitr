require_relative '../uri/postgres'
require 'yaml'
require 'pathname'
require 'securerandom'

module PITR
  module Config
    class Base
      attr_reader :config

      def initialize(path)
        @config = YAML.load_file(path)
      end
    end

    class DB < Base
      def user
        db.fetch('user')
      end

      def host
        db.fetch('host')
      end

      def local_port
        db.fetch('local_port', port)
      end

      def port
        db.fetch('port', URI::Postgres::DEFAULT_PORT)
      end

      def name
        db.fetch('name')
      end

      def version
        db.fetch('version')
      end

      def password
        db.fetch('password')
      end

      def params
        db.fetch('params', {})
      end

      def url
        URI::Postgres.build( components(host, port) )
      end

      def local_url
        URI::Postgres.build( components('localhost', local_port) )
      end

      private

      def db
        config.fetch('db')
      end

      def components(host, port)
        {
          userinfo: [user, password].join(':'),
          host: host,
          port: port,
          path: '/' + name,
          query: query_string,
        }
      end

      def query_string
        return if params.empty?
        params&.map{|kv| kv.join('=') }&.join('&')
      end
    end

    class Minio < Base
      def host
        minio.fetch('host')
      end

      def local_port
        minio.fetch('local_port')
      end

      def port
        minio.fetch('port', 443)
      end

      def access_key
        minio.fetch('access_key')
      end

      def secret_key
        minio.fetch('secret_key')
      end

      private

      def minio
        config.fetch('minio')
      end
    end
  end
end
