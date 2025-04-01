(defcolumns (X :i16) (Y :i16))

(defconstraint test ()
  (== X
     (if (== 0 Y) 0 16)))
