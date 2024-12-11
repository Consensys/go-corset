(defpurefun ((vanishes! :@loob) x) x)

(defcolumns (X :@loob) Y Z)
(defconstraint c1 ()
  (if X
      (begin
       (vanishes! Y)
       (vanishes! Z))))
