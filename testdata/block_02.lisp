(defpurefun ((vanishes! :𝔽@loob) x) x)

(defcolumns (X :i16@loob) Y Z)
(defconstraint c1 ()
  (if X
      (begin
       (vanishes! Y)
       (vanishes! Z))))
