;;error:6:13-14:column is untyped
;;error:6:15-17:empty column declaration
;;error:7:16-22:unknown type
;;error:8:16-23:unknown type
;;error:9:13-30:column is untyped
(defcolumns X ())
(defcolumns (Y :))
(defcolumns (Y :@prove))
(defcolumns (Z :display :hex))
