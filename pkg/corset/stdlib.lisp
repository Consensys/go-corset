;; Copyright Consensys Software Inc.
;;
;; Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with
;; the License. You may obtain a copy of the License at
;;
;; http://www.apache.org/licenses/LICENSE-2.0
;;
;; Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
;; an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
;; specific language governing permissions and limitations under the License.
;;
;; SPDX-License-Identifier: Apache-2.0

(defpurefun (vanishes! e0) (== 0 e0))
;;(defpurefun ((force-bin :binary :force) x) x)

;; =============================================================================
;; Conditionals
;; =============================================================================
(defpurefun (if-zero cond then) (if (== 0 cond) then))
(defpurefun (if-zero cond then else) (if (== 0 cond) then else))
(defpurefun (if-not-zero cond then) (if (!= 0 cond) then))
(defpurefun (if-not-zero cond then else) (if (!= 0 cond) then else))
(defpurefun (if-eq x val then) (if (eq! x val) then))
(defpurefun (if-eq-else x val then else) (if (eq! x val) then else))
(defpurefun (if-not-eq A B then) (if (!= A B) then))

;; =============================================================================
;; Boolean connectives
;; =============================================================================
(defpurefun (or! (a :bool) (b :bool)) (if a (== 0 0) b))
(defpurefun ((eq! :bool) x y) (== x y))

;; =============================================================================
;; Chronological functions
;; =============================================================================
(defpurefun (next X) (shift X 1))
(defpurefun (prev X) (shift X -1))
;; Ensure e0 has increased by offset w.r.t previous row.
(defpurefun (did-inc! e0 offset) (== e0 (+ (prev e0) offset)))
;; Ensure e0 has decreased by offset w.r.t previous row.
(defpurefun (did-dec! e0 offset) (==  e0 (- (prev e0) offset)))
;; Ensure e0 will increase by offset w.r.t next row.
(defpurefun (will-inc! e0 offset) (will-eq! e0 (+ e0 offset)))
;; Ensure e0 will decrease by offset w.r.t next row.
(defpurefun (will-dec! e0 offset) (== (next e0) (- e0 offset)))
;; Ensure e0 remained constant w.r.t previous row.
(defpurefun (remained-constant! e0) (== e0 (prev e0)))
;; Ensure e0 will remain constant w.r.t next row.
(defpurefun (will-remain-constant! e0) (will-eq! e0 e0))
;; Ensure e0 has changed its value w.r.t previous row.
(defpurefun (did-change! e0) (!= e0 (prev e0)))
;; Ensure e0 will remain constant w.r.t next row.
(defpurefun (will-change! e0) (will-neq! e0 e0))
;; Ensure e1 equals value of e0 in previous row.
(defpurefun (was-eq! e0 e1) (== (prev e0) e1))
;; Ensure e1 will equal value of e0 in next row.
(defpurefun (will-eq! e0 e1) (== (next e0) e1))
;; Ensure e1 will not equal value of e0 in next row.
(defpurefun (will-neq! e0 e1) (!= (next e0) e1))

;; =============================================================================
;; Helpers
;; =============================================================================

;; counter constancy constraint
(defpurefun (counter-constancy ct X)
  (if-not-zero ct
               (remained-constant! X)))

;; perspective constancy constraint
(defpurefun (perspective-constancy PERSPECTIVE_SELECTOR X)
            (if-not-zero (* PERSPECTIVE_SELECTOR (prev PERSPECTIVE_SELECTOR))
                         (remained-constant! X)))

;; base-X decomposition constraints
(defpurefun (base-X-decomposition ct base acc digits)
  (if-zero ct
           (== acc digits)
           (== acc (+ (* base (prev acc)) digits))))

;; byte decomposition constraint
(defpurefun (byte-decomposition ct acc bytes) (base-X-decomposition ct 256 acc bytes))

;; bit decomposition constraint
(defpurefun (bit-decomposition ct acc bits) (base-X-decomposition ct 2 acc bits))

;; plateau constraints
(defpurefun (plateau-constraint CT (X :binary) C)
            (begin (debug (stamp-constancy CT C))
                   (if-zero C
                            (== X 1)
                            (if (== CT 0)
                                (vanishes! X)
                              (if (==  CT C)
                                  (did-inc! X 1)
                                (remained-constant! X))))))

;; stamp constancy imposes that the column C may only
;; change at rows where the STAMP column changes.
(defpurefun (stamp-constancy STAMP C)
            (if (will-remain-constant! STAMP)
                (will-remain-constant! C)))
