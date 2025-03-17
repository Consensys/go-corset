(defpurefun (vanishes! x) (== 0 x))

(defcolumns (X :i16) (Y :i16) (Z :i16))
(defconstraint test ()
  (vanishes! (- Z (if (== 0 X) (if (== 0 Y) 0 16)))))
