(defcolumns (X :i16) (Y :i16) (Z :i16))

(defconstraint test ()
  ;; if x == 0 && y == 0 then z == 0
  (if (== 0 X) (== Z (if (== 0 Y) 0))))

(defconstraint test ()
  (if (== 0 X) (== Z (if (== 0 Y) 0 16))))
