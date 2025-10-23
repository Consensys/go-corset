(defcolumns (X :i16) (Y :i16) (Z :i16))
(defconstraint test ()
  (if (== 0 X) (== 0 (- Z (if (== 0 Y) 0 16)))))
