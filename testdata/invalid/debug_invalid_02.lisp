;;error:6:22-31:void expression not permitted here
(defpurefun ((vanishes! :bool :force) x) (== 0 x))
(defcolumns (X :i16) (Y :i16))

(defconstraint c1 ()
  (vanishes! (- X (* (debug X) Y))))
