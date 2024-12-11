(defpurefun ((vanishes! :@loob) x) x)

(defcolumns (X :@loob) (Y :@loob) Z)
(defconstraint test ()
  (if X (vanishes! (- Z (if Y 0)))))

(defconstraint test ()
  (if X (vanishes! (- Z (if Y 0 16)))))
