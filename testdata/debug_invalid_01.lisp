;;error:6:17-26:void expression not permitted here
(defpurefun ((vanishes! :@loob :force) e0) e0)
(defcolumns X Y)

(defconstraint c1 ()
  (vanishes! (- (debug X) Y)))
